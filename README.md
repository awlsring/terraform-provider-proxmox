# Terraform Proxmox Provider


This is a Terraform provider for Proxmox. The API calls made in this repository are done using a generated go client from a [Smithy Proxmox Model](https://github.com/awlsring/ProxmoxModel)

Currently this is very early in development and isn't suitable for use yet. Many things may change.

[Terraform Registry Page](https://registry.terraform.io/providers/awlsring/proxmox/latest)

## Currently supported resources:

#### Data source

* Resource pools
* Network bridges
* Network bonds
* Nodes
* Virtual Machines
* Templates
* Storage class LVM
* Storage class LVM Thinpool
* Storage class NFS
* Storage class ZFS
* Node storage ZFS
* Node storage NFS
* Node storage LVM
* Node storage LVM Thinpool
* LVM Thinpool
* LVM
* ZFS pool

#### Resource

* Resource pool
* Network bond
* Network bridge
* LVM
* LVM Thinpool
* ZFS pool
* Storage class LVM
* Storage class LVM Thinpool
* Storage class NFS
* Storage class ZFS

## Planned resources

### Data source

* Realms
* Users
* Groups

### Resource

* Virtual Machine
* User
* Group