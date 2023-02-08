# Terraform Proxmox

## Dev

Enable logging

```shell
export TF_LOG=DEBUG
```

Store logs
```shell
TF_LOG=TRACE TF_LOG_PATH=trace.txt terraform plan
```