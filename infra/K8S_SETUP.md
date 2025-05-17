# Kubernetes Deployment

latest image of the codebase is pushed to dockerhub at [ahmedkhd36/zendesk-grpc-service](https://hub.docker.com/r/ahmedkhd36/zendesk-grpc-service)

## 1. Deploy PersistentVolumeClaim

in this pvc i use hostpath as storageClass as it comes by default with docker k8s cluster but feel free to change it as you prefered

to make sure the pvc was created successfuly check the status of the pvc you should see smth like "
Waiting for a volume to be created...."

```bash
kubectl apply -f infra/pvc.yaml
```

---

## 2. Copy your `database.db` file into the PersistentVolume

Kubernetes does not natively support copying files into a PVC before a pod is running. The typical approach is:

- Start a temporary pod that mounts the PVC.
- Copy the file into the pod.
- Delete the pod.

Example:

```bash
# Start a temporary pod that mounts the PVC
kubectl run db-copier --image=alpine --restart=Never --overrides='{"spec":{"volumes":[{"name":"db","persistentVolumeClaim":{"claimName":"zendesk-db-pvc"}}],"containers":[{"name":"db-copier","image":"alpine","command":["sleep","3600"],"volumeMounts":[{"mountPath":"/mnt","name":"db"}]}]}}' -- sleep 3600

# In another terminal, copy your database.db into the pod:
kubectl cp ./database.db db-copier:/mnt/database.db

# Delete the temporary pod:
kubectl delete pod db-copier
```

---

## 3. Deploy the Application and Service

```bash
kubectl apply -f infra/deployment.yaml
kubectl apply -f infra/service.yaml
```

---

## 4. Access the Service

- The service is exposed as a ClusterIP by default. To access it externally, you can port-forward:

```
kubectl port-forward svc/zendesk-grpc-service 50051:50051
```

## 5. cleanup 
```bash
kubectl delete -f ./infra && kubectl delete pod db-copier
```
