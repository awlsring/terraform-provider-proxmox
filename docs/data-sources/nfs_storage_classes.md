---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "proxmox_nfs_storage_classes Data Source - terraform-provider-proxmox"
subcategory: ""
description: |-
  
---

# proxmox_nfs_storage_classes (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filters` (Attributes List) (see [below for nested schema](#nestedatt--filters))

### Read-Only

- `nfs_storage_classes` (Attributes List) (see [below for nested schema](#nestedatt--nfs_storage_classes))

<a id="nestedatt--filters"></a>
### Nested Schema for `filters`

Required:

- `name` (String) The name of the attribute to filter on.
- `values` (List of String) The value(s) to be used in the filter.


<a id="nestedatt--nfs_storage_classes"></a>
### Nested Schema for `nfs_storage_classes`

Read-Only:

- `content_types` (List of String) The content types that can be stored on this storage class.
- `export` (String) The remote export path of the NFS server.
- `id` (String) The identifier of the storage class.
- `mount` (String) The local mount of the NFS share that should be implemented by each node.
- `nodes` (List of String) Nodes that implement this storage class.
- `server` (String) The NFS server used in the storage class.


