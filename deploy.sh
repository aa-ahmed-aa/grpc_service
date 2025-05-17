#!/bin/bash

set -e

echo "Applying PersistentVolume and PersistentVolumeClaim..."
kubectl apply -f infra/pvc.yaml

echo "Waiting for PVC to be Bound..."
while [[ $(kubectl get pvc zendesk-db-pvc -o jsonpath='{.status.phase}') != "Bound" ]]; do
  echo "  PVC not bound yet. Waiting 2s..."
  sleep 2
done
echo "PVC is Bound."

echo "Creating temporary pod to copy database.db into the volume..."
kubectl run db-copier --image=alpine --restart=Never --overrides='{"spec":{"volumes":[{"name":"db","persistentVolumeClaim":{"claimName":"zendesk-db-pvc"}}],"containers":[{"name":"db-copier","image":"alpine","command":["sleep","3600"],"volumeMounts":[{"mountPath":"/mnt","name":"db"}]}]}}' -- sleep 3600

echo "Waiting for db-copier pod to be Running..."
while [[ $(kubectl get pod db-copier -o jsonpath='{.status.phase}') != "Running" ]]; do
  echo "  db-copier pod not running yet. Waiting 2s..."
  sleep 2
done

echo "Copying database.db into the pod volume..."
kubectl cp ./database.db db-copier:/mnt/database.db

echo "Deleting db-copier pod..."
kubectl delete pod db-copier

echo "Deploying application and service..."
kubectl apply -f infra/deployment.yaml
kubectl apply -f infra/service.yaml

echo "Waiting for zendesk-grpc-service pod to be Running..."
while [[ $(kubectl get pods -l app=zendesk-grpc-service -o jsonpath='{.items[0].status.phase}') != "Running" ]]; do
  echo "  App pod not running yet. Waiting 2s..."
  sleep 2
done

echo "Port-forwarding service to localhost:50051..."
kubectl port-forward svc/zendesk-grpc-service 50051:50051 &
PF_PID=$!

# Wait a few seconds for port-forward to establish
sleep 3

echo "Checking if port 50051 is open..."
if nc -z localhost 50051; then
  echo "✅ App is live and exposed on localhost:50051"
else
  echo "❌ App is not responding on localhost:50051"
fi

echo "To stop port-forwarding, run: kill $PF_PID"