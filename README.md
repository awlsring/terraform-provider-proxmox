# Terraform Proxmox Provider

This is a Terraform provider for Proxmox. The API calls made in this repository are done using a generated go client from a [Smithy Proxmox Model](https://github.com/awlsring/ProxmoxModel)

Currently this is very early in development and isn't suitable for use yet. Many things may change.

## Currently supported resources:

#### Data source

* Resource pools
* Network bridges
* Network bonds
* Nodes
* Virtual Machines
* Templates
* Storage

#### Resource

* Resource pool

## Planned resources

### Data source

* Realms
* Users
* Groups

### Resource

* Network bridge
* Network bond
* Virtual Machine
* Storage
* User
* Group