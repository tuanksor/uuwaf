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
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Configure logging to stdout
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get MySQL connection details from environment variables
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbName := os.Getenv("MYSQL_DATABASE")

	if dbUser == "" || dbPass == "" || dbHost == "" || dbPort == "" || dbName == "" {
		dbUser = "<user>"
		dbPass = "<password>"
		dbHost = "<ip or domain>"
		dbPort = "3306"
		dbName = "uuwaf"
	}

	// Construct MySQL connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to MariaDB database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Connect to Kubernetes API
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/kubernetes/admin.conf")
	if err != nil {
		log.Fatalf("Failed to get Kubernetes config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	// Get current pod namespace
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "uusec" // Use default namespace if POD_NAMESPACE is not set
	}

	// Variable to track the last processed audit ID
	lastID := 0
	log.Printf("lastID: %v", lastID)
	log.Printf("Waiting new event")
	for {
		// Create a context for the current iteration
		ctx := context.Background()

		// Query for the latest audit record
		row := db.QueryRow("SELECT id, type, info FROM waf_audits WHERE id > ? ORDER BY id DESC LIMIT 1", lastID)

		var id int
		var auditType, info string
		err = row.Scan(&id, &auditType, &info)

		if err == nil {
			// New audit record found
			if auditType == "Site" || auditType == "Rule" {
				log.Printf("New event found: ID=%d, Info=%s", id, info)

				// Get list of uuwaf pods
				pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
					LabelSelector: "app=uuwaf", // Replace with your actual label selector if needed
				})
				if err != nil {
					log.Printf("Failed to get pod list: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Execute uuwaf -s reload command on each pod using Go API client
				for _, pod := range pods.Items {
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

					exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
					if err != nil {
						log.Printf("Failed to create SPDY executor for pod %s: %v", pod.Name, err)
						continue
					}

					var stdout, stderr string
					outbuf := &bytes.Buffer{}
					errbuf := &bytes.Buffer{}

					err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
						Stdout: outbuf,
						Stderr: errbuf,
					})
					stdout = outbuf.String()
					stderr = errbuf.String()

					if err != nil {
						log.Printf("Failed to reflect configuration on pod %s: %v, stdout: %s, stderr: %s", pod.Name, err, stdout, stderr)
					} else {
						log.Printf("Successfully reflect configuration on pod %s, stdout: %s", pod.Name, stdout)
					}
				}
			}
			// Update lastID
			lastID = id
      log.Printf("Updated lastID: %v", lastID)
		} else if err != sql.ErrNoRows {
			log.Printf("Error querying events: %v", err)
		}

		// Wait for 5 seconds before checking again
		time.Sleep(5 * time.Second)
	}
}
