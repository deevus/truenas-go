# TrueNAS API Feature Matrix

TrueNAS version: 25.04

Total API methods: 771 | Implemented: 84 (10.9%) | Tested: 84 (100.0% of implemented)

## Covered Namespaces

| Go Service | Namespaces | API Methods | Implemented | Tested |
|------------|------------|:-----------:|:-----------:|:------:|
| AppService | app, app.image, app.registry | 38 | 15 (39%) | 15 (100%) |
| CloudSyncService | cloudsync, cloudsync.credentials | 20 | 9 (45%) | 9 (100%) |
| CronService | cronjob | 6 | 6 (100%) | 6 (100%) |
| DatasetService | pool, pool.dataset | 50 | 5 (10%) | 5 (100%) |
| DockerService | docker | 8 | 2 (25%) | 2 (100%) |
| FilesystemService | filesystem | 13 | 2 (15%) | 2 (100%) |
| GroupService | group | 8 | 5 (62%) | 5 (100%) |
| InterfaceService | interface | 23 | 1 (4%) | 1 (100%) |
| NetworkService | network.general | 1 | 1 (100%) | 1 (100%) |
| ReportingService | reporting | 8 | 2 (25%) | 2 (100%) |
| SnapshotService | zfs.snapshot | 9 | 7 (78%) | 7 (100%) |
| SystemService | system | 14 | 2 (14%) | 2 (100%) |
| UserService | user | 13 | 5 (38%) | 5 (100%) |
| VMService | vm, vm.device | 51 | 10 (20%) | 10 (100%) |
| VirtService | virt.global, virt.instance | 18 | 12 (67%) | 12 (100%) |

### AppService — `app` (28 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| app.available |  |  |  |  |
| app.available_space | ✓ | AvailableSpace | ✓ | 2 |
| app.categories |  |  |  |  |
| app.certificate_authority_choices |  |  |  |  |
| app.certificate_choices |  |  |  |  |
| app.config |  |  |  |  |
| app.container_console_choices |  |  |  |  |
| app.container_ids |  |  |  |  |
| app.convert_to_custom |  |  |  |  |
| app.create | ✓ | CreateApp | ✓ | 5 |
| app.delete | ✓ | DeleteApp | ✓ | 2 |
| app.get_instance |  |  |  |  |
| app.gpu_choices |  |  |  |  |
| app.ip_choices |  |  |  |  |
| app.latest |  |  |  |  |
| app.outdated_docker_images |  |  |  |  |
| app.pull_images |  |  |  |  |
| app.query | ✓ | GetApp, GetAppWithConfig, ListApps | ✓ | 12 |
| app.redeploy | ✓ | RedeployApp | ✓ | 2 |
| app.rollback |  |  |  |  |
| app.rollback_versions |  |  |  |  |
| app.similar |  |  |  |  |
| app.start | ✓ | StartApp | ✓ | 2 |
| app.stop | ✓ | StopApp | ✓ | 2 |
| app.update | ✓ | UpdateApp | ✓ | 5 |
| app.upgrade | ✓ | UpgradeApp | ✓ | 2 |
| app.upgrade_summary | ✓ | UpgradeSummary | ✓ | 2 |
| app.used_ports |  |  |  |  |

### AppService — `app.image` (5 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| app.image.delete |  |  |  |  |
| app.image.dockerhub_rate_limit |  |  |  |  |
| app.image.get_instance |  |  |  |  |
| app.image.pull |  |  |  |  |
| app.image.query | ✓ | ListImages | ✓ | 3 |

### AppService — `app.registry` (5 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| app.registry.create | ✓ | CreateRegistry | ✓ | 6 |
| app.registry.delete | ✓ | DeleteRegistry | ✓ | 2 |
| app.registry.get_instance |  |  |  |  |
| app.registry.query | ✓ | GetRegistry, ListRegistries | ✓ | 9 |
| app.registry.update | ✓ | UpdateRegistry | ✓ | 4 |

### CloudSyncService — `cloudsync` (14 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| cloudsync.abort |  |  |  |  |
| cloudsync.create | ✓ | CreateTask | ✓ | 3 |
| cloudsync.create_bucket |  |  |  |  |
| cloudsync.delete | ✓ | DeleteTask | ✓ | 2 |
| cloudsync.get_instance |  |  |  |  |
| cloudsync.list_buckets |  |  |  |  |
| cloudsync.list_directory |  |  |  |  |
| cloudsync.onedrive_list_drives |  |  |  |  |
| cloudsync.providers |  |  |  |  |
| cloudsync.query | ✓ | GetTask, ListTasks | ✓ | 9 |
| cloudsync.restore |  |  |  |  |
| cloudsync.sync | ✓ | Sync | ✓ | 2 |
| cloudsync.sync_onetime |  |  |  |  |
| cloudsync.update | ✓ | UpdateTask | ✓ | 2 |

### CloudSyncService — `cloudsync.credentials` (6 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| cloudsync.credentials.create | ✓ | CreateCredential | ✓ | 4 |
| cloudsync.credentials.delete | ✓ | DeleteCredential | ✓ | 2 |
| cloudsync.credentials.get_instance |  |  |  |  |
| cloudsync.credentials.query | ✓ | GetCredential, ListCredentials | ✓ | 10 |
| cloudsync.credentials.update | ✓ | UpdateCredential | ✓ | 3 |
| cloudsync.credentials.verify |  |  |  |  |

### CronService — `cronjob` (6 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| cronjob.create | ✓ | Create | ✓ | 3 |
| cronjob.delete | ✓ | Delete | ✓ | 2 |
| cronjob.get_instance | ✓ | Get | ✓ | 3 |
| cronjob.query | ✓ | List | ✓ | 3 |
| cronjob.run | ✓ | Run | ✓ | 3 |
| cronjob.update | ✓ | Update | ✓ | 2 |

### DatasetService — `pool` (24 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| pool.attach |  |  |  |  |
| pool.attachments |  |  |  |  |
| pool.create |  |  |  |  |
| pool.ddt_prefetch |  |  |  |  |
| pool.ddt_prune |  |  |  |  |
| pool.detach |  |  |  |  |
| pool.expand |  |  |  |  |
| pool.export |  |  |  |  |
| pool.filesystem_choices |  |  |  |  |
| pool.get_disks |  |  |  |  |
| pool.get_instance |  |  |  |  |
| pool.import_find |  |  |  |  |
| pool.import_pool |  |  |  |  |
| pool.is_upgraded |  |  |  |  |
| pool.offline |  |  |  |  |
| pool.online |  |  |  |  |
| pool.processes |  |  |  |  |
| pool.query | ✓ | ListPools | ✓ | 4 |
| pool.remove |  |  |  |  |
| pool.replace |  |  |  |  |
| pool.scrub |  |  |  |  |
| pool.update |  |  |  |  |
| pool.upgrade |  |  |  |  |
| pool.validate_name |  |  |  |  |

### DatasetService — `pool.dataset` (26 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| pool.dataset.attachments |  |  |  |  |
| pool.dataset.change_key |  |  |  |  |
| pool.dataset.checksum_choices |  |  |  |  |
| pool.dataset.compression_choices |  |  |  |  |
| pool.dataset.create | ✓ | CreateDataset, CreateZvol | ✓ | 8 |
| pool.dataset.delete | ✓ | DeleteDataset, DeleteZvol | ✓ | 5 |
| pool.dataset.destroy_snapshots |  |  |  |  |
| pool.dataset.details |  |  |  |  |
| pool.dataset.encryption_algorithm_choices |  |  |  |  |
| pool.dataset.encryption_summary |  |  |  |  |
| pool.dataset.export_key |  |  |  |  |
| pool.dataset.export_keys |  |  |  |  |
| pool.dataset.export_keys_for_replication |  |  |  |  |
| pool.dataset.get_instance |  |  |  |  |
| pool.dataset.get_quota |  |  |  |  |
| pool.dataset.inherit_parent_encryption_properties |  |  |  |  |
| pool.dataset.lock |  |  |  |  |
| pool.dataset.processes |  |  |  |  |
| pool.dataset.promote |  |  |  |  |
| pool.dataset.query | ✓ | GetDataset, ListDatasets, GetZvol | ✓ | 13 |
| pool.dataset.recommended_zvol_blocksize |  |  |  |  |
| pool.dataset.recordsize_choices |  |  |  |  |
| pool.dataset.set_quota |  |  |  |  |
| pool.dataset.snapshot_count |  |  |  |  |
| pool.dataset.unlock |  |  |  |  |
| pool.dataset.update | ✓ | UpdateDataset, UpdateZvol | ✓ | 7 |

### DockerService — `docker` (8 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| docker.backup |  |  |  |  |
| docker.config | ✓ | GetConfig | ✓ | 3 |
| docker.delete_backup |  |  |  |  |
| docker.list_backups |  |  |  |  |
| docker.nvidia_present |  |  |  |  |
| docker.restore_backup |  |  |  |  |
| docker.status | ✓ | GetStatus | ✓ | 3 |
| docker.update |  |  |  |  |

### FilesystemService — `filesystem` (13 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| filesystem.can_access_as_user |  |  |  |  |
| filesystem.chown |  |  |  |  |
| filesystem.get |  |  |  |  |
| filesystem.get_zfs_attributes |  |  |  |  |
| filesystem.getacl |  |  |  |  |
| filesystem.listdir |  |  |  |  |
| filesystem.mkdir |  |  |  |  |
| filesystem.put |  |  |  |  |
| filesystem.set_zfs_attributes |  |  |  |  |
| filesystem.setacl |  |  |  |  |
| filesystem.setperm | ✓ | SetPermissions | ✓ | 4 |
| filesystem.stat | ✓ | Stat | ✓ | 4 |
| filesystem.statfs |  |  |  |  |

### GroupService — `group` (8 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| group.create | ✓ | Create | ✓ | 3 |
| group.delete | ✓ | Delete | ✓ | 1 |
| group.get_group_obj |  |  |  |  |
| group.get_instance | ✓ | Get | ✓ | 4 |
| group.get_next_gid |  |  |  |  |
| group.has_password_enabled_user |  |  |  |  |
| group.query | ✓ | List | ✓ | 3 |
| group.update | ✓ | Update | ✓ | 2 |

### InterfaceService — `interface` (23 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| interface.bridge_members_choices |  |  |  |  |
| interface.cancel_rollback |  |  |  |  |
| interface.checkin |  |  |  |  |
| interface.checkin_waiting |  |  |  |  |
| interface.choices |  |  |  |  |
| interface.commit |  |  |  |  |
| interface.create |  |  |  |  |
| interface.default_route_will_be_removed |  |  |  |  |
| interface.delete |  |  |  |  |
| interface.get_instance |  |  |  |  |
| interface.has_pending_changes |  |  |  |  |
| interface.ip_in_use |  |  |  |  |
| interface.lacpdu_rate_choices |  |  |  |  |
| interface.lag_ports_choices |  |  |  |  |
| interface.query | ✓ | List, Get | ✓ | 8 |
| interface.rollback |  |  |  |  |
| interface.save_default_route |  |  |  |  |
| interface.services_restarted_on_sync |  |  |  |  |
| interface.update |  |  |  |  |
| interface.vlan_parent_interface_choices |  |  |  |  |
| interface.websocket_interface |  |  |  |  |
| interface.websocket_local_ip |  |  |  |  |
| interface.xmit_hash_policy_choices |  |  |  |  |

### NetworkService — `network.general` (1 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| network.general.summary | ✓ | GetSummary | ✓ | 5 |

### ReportingService — `reporting` (8 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| reporting.config |  |  |  |  |
| reporting.get_data |  |  |  |  |
| reporting.graph |  |  |  |  |
| reporting.graphs |  |  |  |  |
| reporting.netdata_get_data | ✓ | GetData | ✓ | 4 |
| reporting.netdata_graph |  |  |  |  |
| reporting.netdata_graphs | ✓ | ListGraphs | ✓ | 4 |
| reporting.update |  |  |  |  |

### SnapshotService — `zfs.snapshot` (9 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| zfs.snapshot.clone | ✓ | Clone | ✓ | 3 |
| zfs.snapshot.create | ✓ | Create | ✓ | 6 |
| zfs.snapshot.delete | ✓ | Delete | ✓ | 2 |
| zfs.snapshot.get_instance |  |  |  |  |
| zfs.snapshot.hold | ✓ | Hold | ✓ | 2 |
| zfs.snapshot.query | ✓ | Get, List, Query | ✓ | 15 |
| zfs.snapshot.release | ✓ | Release | ✓ | 2 |
| zfs.snapshot.rollback | ✓ | Rollback | ✓ | 3 |
| zfs.snapshot.update |  |  |  |  |

### SystemService — `system` (14 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| system.boot_id |  |  |  |  |
| system.debug |  |  |  |  |
| system.feature_enabled |  |  |  |  |
| system.host_id |  |  |  |  |
| system.info | ✓ | GetInfo | ✓ | 3 |
| system.license_update |  |  |  |  |
| system.product_type |  |  |  |  |
| system.ready |  |  |  |  |
| system.reboot |  |  |  |  |
| system.release_notes_url |  |  |  |  |
| system.shutdown |  |  |  |  |
| system.state |  |  |  |  |
| system.version | ✓ | GetVersion | ✓ | 3 |
| system.version_short |  |  |  |  |

### UserService — `user` (13 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| user.create | ✓ | Create | ✓ | 3 |
| user.delete | ✓ | Delete | ✓ | 1 |
| user.get_instance | ✓ | Get | ✓ | 4 |
| user.get_next_uid |  |  |  |  |
| user.get_user_obj |  |  |  |  |
| user.has_local_administrator_set_up |  |  |  |  |
| user.query | ✓ | List | ✓ | 3 |
| user.renew_2fa_secret |  |  |  |  |
| user.set_password |  |  |  |  |
| user.setup_local_administrator |  |  |  |  |
| user.shell_choices |  |  |  |  |
| user.unset_2fa_secret |  |  |  |  |
| user.update | ✓ | Update | ✓ | 2 |

### VMService — `vm` (35 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| vm.bootloader_options |  |  |  |  |
| vm.bootloader_ovmf_choices |  |  |  |  |
| vm.clone |  |  |  |  |
| vm.cpu_model_choices |  |  |  |  |
| vm.create | ✓ | CreateVM | ✓ | 5 |
| vm.delete | ✓ | DeleteVM | ✓ | 2 |
| vm.export_disk_image |  |  |  |  |
| vm.flags |  |  |  |  |
| vm.get_available_memory |  |  |  |  |
| vm.get_console |  |  |  |  |
| vm.get_display_devices |  |  |  |  |
| vm.get_display_web_uri |  |  |  |  |
| vm.get_instance | ✓ | GetVM | ✓ | 5 |
| vm.get_memory_usage |  |  |  |  |
| vm.get_vm_memory_info |  |  |  |  |
| vm.get_vmemory_in_use |  |  |  |  |
| vm.guest_architecture_and_machine_choices |  |  |  |  |
| vm.import_disk_image |  |  |  |  |
| vm.log_file_download |  |  |  |  |
| vm.log_file_path |  |  |  |  |
| vm.maximum_supported_vcpus |  |  |  |  |
| vm.port_wizard |  |  |  |  |
| vm.poweroff |  |  |  |  |
| vm.query |  |  |  |  |
| vm.random_mac |  |  |  |  |
| vm.resolution_choices |  |  |  |  |
| vm.restart |  |  |  |  |
| vm.resume |  |  |  |  |
| vm.start | ✓ | StartVM | ✓ | 2 |
| vm.status |  |  |  |  |
| vm.stop | ✓ | StopVM | ✓ | 3 |
| vm.supports_virtualization |  |  |  |  |
| vm.suspend |  |  |  |  |
| vm.update | ✓ | UpdateVM | ✓ | 3 |
| vm.virtualization_details |  |  |  |  |

### VMService — `vm.device` (16 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| vm.device.bind_choices |  |  |  |  |
| vm.device.create | ✓ | CreateDevice | ✓ | 10 |
| vm.device.delete | ✓ | DeleteDevice | ✓ | 2 |
| vm.device.disk_choices |  |  |  |  |
| vm.device.get_instance |  |  |  |  |
| vm.device.iommu_enabled |  |  |  |  |
| vm.device.iotype_choices |  |  |  |  |
| vm.device.nic_attach_choices |  |  |  |  |
| vm.device.passthrough_device |  |  |  |  |
| vm.device.passthrough_device_choices |  |  |  |  |
| vm.device.pptdev_choices |  |  |  |  |
| vm.device.query | ✓ | ListDevices, GetDevice | ✓ | 8 |
| vm.device.update | ✓ | UpdateDevice | ✓ | 3 |
| vm.device.usb_controller_choices |  |  |  |  |
| vm.device.usb_passthrough_choices |  |  |  |  |
| vm.device.usb_passthrough_device |  |  |  |  |

### VirtService — `virt.global` (5 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| virt.global.bridge_choices |  |  |  |  |
| virt.global.config | ✓ | GetGlobalConfig | ✓ | 6 |
| virt.global.get_network |  |  |  |  |
| virt.global.pool_choices |  |  |  |  |
| virt.global.update | ✓ | UpdateGlobalConfig | ✓ | 3 |

### VirtService — `virt.instance` (13 methods)

| API Method | Implemented | Go Method | Tested | Tests |
|------------|:-----------:|-----------|:------:|------:|
| virt.instance.create | ✓ | CreateInstance | ✓ | 3 |
| virt.instance.delete | ✓ | DeleteInstance | ✓ | 2 |
| virt.instance.device_add | ✓ | AddDevice | ✓ | 4 |
| virt.instance.device_delete | ✓ | DeleteDevice | ✓ | 2 |
| virt.instance.device_list | ✓ | ListDevices | ✓ | 4 |
| virt.instance.device_update |  |  |  |  |
| virt.instance.get_instance | ✓ | GetInstance | ✓ | 4 |
| virt.instance.image_choices |  |  |  |  |
| virt.instance.query | ✓ | ListInstances | ✓ | 5 |
| virt.instance.restart |  |  |  |  |
| virt.instance.start | ✓ | StartInstance | ✓ | 2 |
| virt.instance.stop | ✓ | StopInstance | ✓ | 3 |
| virt.instance.update | ✓ | UpdateInstance | ✓ | 3 |

## Uncovered Namespaces (93 namespaces, 491 methods)

| Namespace | Methods |
|-----------|--------:|
| acme.dns.authenticator | 6 |
| activedirectory | 5 |
| alert | 5 |
| alertclasses | 2 |
| alertservice | 6 |
| api_key | 6 |
| app.ix_volume | 2 |
| audit | 5 |
| auth | 14 |
| auth.twofactor | 2 |
| boot | 7 |
| boot.environment | 5 |
| catalog | 6 |
| certificate | 9 |
| certificateauthority | 6 |
| cloud_backup | 12 |
| config | 3 |
| core | 14 |
| device | 1 |
| directoryservices | 3 |
| disk | 13 |
| dns | 1 |
| docker.network | 2 |
| enclosure.label | 1 |
| failover | 16 |
| failover.disabled | 1 |
| failover.reboot | 2 |
| filesystem.acltemplate | 6 |
| ftp | 2 |
| hardware.memory | 1 |
| idmap | 8 |
| initshutdownscript | 5 |
| interface.capabilities | 2 |
| ipmi | 1 |
| ipmi.chassis | 2 |
| ipmi.lan | 4 |
| ipmi.mc | 1 |
| ipmi.sel | 3 |
| ipmi.sensors | 1 |
| iscsi.auth | 5 |
| iscsi.extent | 6 |
| iscsi.global | 6 |
| iscsi.initiator | 5 |
| iscsi.portal | 6 |
| iscsi.target | 6 |
| iscsi.targetextent | 5 |
| jbof | 7 |
| k8s_to_docker | 2 |
| kerberos | 2 |
| kerberos.keytab | 5 |
| kerberos.realm | 5 |
| keychaincredential | 10 |
| kmip | 5 |
| ldap | 4 |
| mail | 4 |
| network.configuration | 3 |
| nfs | 6 |
| pool.resilver | 2 |
| pool.scrub | 7 |
| pool.snapshottask | 10 |
| privilege | 6 |
| replication | 13 |
| replication.config | 2 |
| reporting.exporters | 6 |
| route | 2 |
| rsynctask | 6 |
| service | 9 |
| sharing.nfs | 5 |
| sharing.smb | 8 |
| smart | 2 |
| smart.test | 10 |
| smb | 6 |
| snmp | 2 |
| ssh | 3 |
| staticroute | 5 |
| support | 9 |
| system.advanced | 10 |
| system.general | 14 |
| system.global | 1 |
| system.ntpserver | 5 |
| system.reboot | 1 |
| system.security | 2 |
| system.security.info | 2 |
| systemdataset | 3 |
| tn_connect | 5 |
| truecommand | 2 |
| truenas | 8 |
| tunable | 6 |
| update | 10 |
| ups | 4 |
| virt.device | 7 |
| virt.volume | 7 |
| vmware | 8 |

## Go Methods Not in API Schema (13 methods)

These Go methods call API endpoints not present in the 25.04 method schema
(e.g., subscription/event channels, version-specific aliases).

| Go Service | Go Method | API Method |
|------------|-----------|------------|
| AppService | SubscribeContainerLogs | app.container_log_follow |
| AppService | SubscribeStats | app.stats |
| FilesystemService | WriteFile | filesystem.file_receive |
| ReportingService | SubscribeRealtime | reporting.realtime |
| SnapshotService | Clone | pool.snapshot.clone |
| SnapshotService | Create | pool.snapshot.create |
| SnapshotService | Delete | pool.snapshot.delete |
| SnapshotService | Hold | pool.snapshot.hold |
| SnapshotService | List | pool.snapshot.query |
| SnapshotService | Query | pool.snapshot.query |
| SnapshotService | Get | pool.snapshot.query |
| SnapshotService | Release | pool.snapshot.release |
| SnapshotService | Rollback | pool.snapshot.rollback |

