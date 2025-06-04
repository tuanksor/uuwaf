/*
 * Author: Nguyen Anh Tuan
 * Email: tuan.gli@gmail.com
 * Description: UUWAF Operator - A service that monitors UUWAF audit database and automatically reloads configurations
 */

package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	corev1 "k8s.io/api/core/v1"

	"operator-uuwaf/config"
	"operator-uuwaf/metrics"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load configuration
	cfg := config.NewDefaultConfig()
	cfg.Database.User = os.Getenv("MYSQL_USER")
	cfg.Database.Password = os.Getenv("MYSQL_PASSWORD")
	cfg.Database.Host = os.Getenv("MYSQL_HOST")
	cfg.Database.Port = os.Getenv("MYSQL_PORT")
	cfg.Database.Name = os.Getenv("MYSQL_DATABASE")
	cfg.Kubernetes.Namespace = os.Getenv("POD_NAMESPACE")

	if err := cfg.Validate(); err != nil {
		sugar.Fatalw("Invalid configuration", "error", err)
	}

	// Create a context that will be canceled on SIGTERM or SIGINT
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		sugar.Infow("Received signal, initiating shutdown", "signal", sig)
		cancel()
	}()

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":8080", nil); err != nil {
			sugar.Errorw("Failed to start metrics server", "error", err)
		}
	}()

	// Connect to MariaDB database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		sugar.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.Pool.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.Pool.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.Pool.ConnMaxLifetime)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		sugar.Fatalw("Failed to ping database", "error", err)
	}

	// Connect to Kubernetes API
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", cfg.Kubernetes.ConfigPath)
	if err != nil {
		sugar.Fatalw("Failed to get Kubernetes config", "error", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		sugar.Fatalw("Failed to create Kubernetes clientset", "error", err)
	}

	// Variable to track the last processed audit ID
	lastID := 0
	sugar.Infow("Starting monitoring", "lastID", lastID)

	// Main loop
	for {
		select {
		case <-ctx.Done():
			sugar.Info("Shutting down gracefully...")
			return
		default:
			start := time.Now()
			if err := processAuditEvents(ctx, db, clientset, k8sConfig, cfg.Kubernetes.Namespace, &lastID, sugar); err != nil {
				sugar.Errorw("Error processing audit events", "error", err)
			}
			metrics.ProcessingDuration.Observe(time.Since(start).Seconds())
			time.Sleep(cfg.Operator.PollInterval)
		}
	}
}

func processAuditEvents(ctx context.Context, db *sql.DB, clientset *kubernetes.Clientset, k8sConfig *rest.Config, namespace string, lastID *int, sugar *zap.SugaredLogger) error {
	// Query for the latest audit record
	row := db.QueryRowContext(ctx, "SELECT id, type, info FROM waf_audits WHERE id > ? ORDER BY id DESC LIMIT 1", *lastID)

	var id int
	var auditType, info string
	err := row.Scan(&id, &auditType, &info)

	if err == nil {
		// New audit record found
		if auditType == "Site" || auditType == "Rule" {
			sugar.Infow("New event found", "id", id, "type", auditType, "info", info)
			metrics.AuditEventsProcessed.WithLabelValues(auditType).Inc()

			// Get list of uuwaf pods
			pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: "app=uuwaf",
			})
			if err != nil {
				return fmt.Errorf("failed to get pod list: %v", err)
			}

			// Execute uuwaf -s reload command on each pod
			for _, pod := range pods.Items {
				if err := reloadPodConfig(ctx, clientset, k8sConfig, pod, namespace, sugar); err != nil {
					metrics.PodReloads.WithLabelValues(pod.Name, "failed").Inc()
					sugar.Errorw("Failed to reload pod", "pod", pod.Name, "error", err)
				} else {
					metrics.PodReloads.WithLabelValues(pod.Name, "success").Inc()
				}
			}
		}
		// Update lastID
		*lastID = id
		metrics.LastProcessedID.Set(float64(*lastID))
		sugar.Infow("Updated lastID", "id", *lastID)
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("error querying events: %v", err)
	}

	return nil
}

func reloadPodConfig(ctx context.Context, clientset *kubernetes.Clientset, k8sConfig *rest.Config, pod corev1.Pod, namespace string, sugar *zap.SugaredLogger) error {
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"/uuwaf/sbin/uuwaf", "-s", "reload"},
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k8sConfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create SPDY executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return fmt.Errorf("failed to execute command: %v, stdout: %s, stderr: %s", err, stdout.String(), stderr.String())
	}

	sugar.Infow("Successfully reloaded configuration", "pod", pod.Name, "stdout", stdout.String())
	return nil
}
