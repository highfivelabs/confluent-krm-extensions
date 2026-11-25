# confluent-krm-extensions

Kubernetes kinds for the Confluent Cloud configuration that Terraform's
`confluentinc/confluent` provider manages and Confluent Cloud's operators do not. Confluent for Kubernetes (CFK) deploys Confluent Platform and manages topics, schemas, connectors, role bindings, cluster links and schema exporters; some of these can point at Confluent Cloud. It does not manage Cloud environments, clusters, networking, service accounts or API keys.

By the numbers: `confluentinc/confluent` is the #18 Terraform provider by downloads
(620,704,761); version 2.95.0 has 70 resources. Operators
cover 6. This repository adds **62 kinds** standing in for
**63 more** of the other 64. Each kind is listed below with its Terraform
equivalent; the one left is under "Not covered".

**Two kinds of host.** Most kinds call the control plane at
api.confluent.cloud. The Schema Registry, Stream Catalog, Kafka REST and
Flink SQL kinds call a data-plane endpoint of their own (see "Data-plane kinds").

**Readiness.** Every kind passes create, update, drift repair and delete
against a fake built from the same descriptors (`TestEveryKindOnFake`).
The kinds marked `live` in `zz_generated.report.md` have also passed them
against the real API (`hack/live/records.json`); treat the rest, including
every kind that needs a Dedicated or Enterprise cluster or private
networking, as untested there until they have.

## How it works

The kinds are generated, not hand-written. `hack/gen/vendor.yaml` pins
Confluent Cloud's OpenAPI spec; [krm-rest-runtime](https://github.com/highfivelabs/krm-rest-runtime)'s
generator finds every resource the API exposes in a create/read/update/delete
shape, keeps those that stand in for a Terraform resource in the gap list
(`hack/gen/terraform-gaps.json`), and writes:

- `types.yaml`: per kind, its API calls, path parameters and Terraform equivalents
- `config/crds/`: one CRD per kind, its spec the API's request body
- `zz_generated.report.md`: the full list

One generic manager (`cmd/manager`) runs every kind on krm-rest-runtime's
reconciler: it creates the resource, reads it back every 10 minutes and
puts back anything changed outside Kubernetes, and deletes it with the
object. Spec fields keep Confluent Cloud's own names, so the API reference
applies field for field. Path parameters are spec fields too.

Hand corrections to the automatic matching (a Terraform type matched to
the wrong API path, or not matched) are in `vendor.yaml`'s `overrides`.

## Install

```sh
kubectl kustomize config/default | KO_DOCKER_REPO=<registry> ko resolve -f - | kubectl apply --server-side -f -
```

Then, in each namespace that manages Confluent Cloud, a Secret with
`apiKey`, `apiSecret` and a Connection named `default`
(`config/samples/connection.yaml`). Objects use it unless
`spec.connectionRef.name` names another. Secret values (tokens, passwords,
secret values) go in a Kubernetes Secret, not the spec:
`spec.secretRefs: {<field>: {name, key}}` fills that body field.

To run only some API groups, install those groups' CRDs from `config/crds`
and pass `--groups` to the manager.

## Kinds

| Kind | Terraform equivalent | Shape | Read |
| --- | --- | --- | --- |
| `access.confluent.hfl.dev/v1alpha1` `AccessPoint` | `confluent_access_point` | crud | `GET /networking/v1/access-points/{id}` |
| `api.confluent.hfl.dev/v1alpha1` `ApiKey` | `confluent_api_key` | crud | `GET /iam/v2/api-keys/{id}` |
| `business.confluent.hfl.dev/v1alpha1` `BusinessMetadata` | `confluent_business_metadata` | crud | `GET /catalog/v1/types/businessmetadatadefs/{bmName}` |
| `business.confluent.hfl.dev/v1alpha1` `BusinessMetadataBinding` | `confluent_business_metadata_binding` | crud | `GET /catalog/v1/entity/type/{typeName}/name/{qualifiedName}/businessmetadata` |
| `byok.confluent.hfl.dev/v1alpha1` `ByokKey` | `confluent_byok_key` | crud | `GET /byok/v1/keys/{id}` |
| `catalog.confluent.hfl.dev/v1alpha1` `CatalogEntityAttributes` | `confluent_catalog_entity_attributes` | settings | `GET /catalog/v1/entity/type/{typeName}/name/{qualifiedName}` |
| `catalog.confluent.hfl.dev/v1alpha1` `CatalogIntegration` | `confluent_catalog_integration` | crud | `GET /tableflow/v1/catalog-integrations/{id}` |
| `certificate.confluent.hfl.dev/v1alpha1` `CertificateAuthority` | `confluent_certificate_authority` | crud | `GET /iam/v2/certificate-authorities/{id}` |
| `certificate.confluent.hfl.dev/v1alpha1` `CertificatePool` | `confluent_certificate_pool` | crud | `GET /iam/v2/certificate-authorities/{certificate_authority_id}/identity-pools/{id}` |
| `connect.confluent.hfl.dev/v1alpha1` `ConnectArtifact` | `confluent_connect_artifact` | crud | `GET /cam/v1/connect-artifacts/{id}` |
| `custom.confluent.hfl.dev/v1alpha1` `CustomConnectorPlugin` | `confluent_custom_connector_plugin` | crud | `GET /connect/v1/custom-connector-plugins/{id}` |
| `custom.confluent.hfl.dev/v1alpha1` `CustomConnectorPluginVersion` | `confluent_custom_connector_plugin_version` | crud | `GET /ccpm/v1/plugins/{plugin_id}/versions/{id}` |
| `dns.confluent.hfl.dev/v1alpha1` `DnsForwarder` | `confluent_dns_forwarder` | crud | `GET /networking/v1/dns-forwarders/{id}` |
| `dns.confluent.hfl.dev/v1alpha1` `DnsRecord` | `confluent_dns_record` | crud | `GET /networking/v1/dns-records/{id}` |
| `environment.confluent.hfl.dev/v1alpha1` `Environment` | `confluent_environment` | crud | `GET /org/v2/environments/{id}` |
| `flink.confluent.hfl.dev/v1alpha1` `FlinkArtifact` | `confluent_flink_artifact` | crud | `GET /artifact/v1/flink-artifacts/{id}` |
| `flink.confluent.hfl.dev/v1alpha1` `FlinkComputePool` | `confluent_flink_compute_pool` | crud | `GET /fcpm/v2/compute-pools/{id}` |
| `flink.confluent.hfl.dev/v1alpha1` `FlinkComputePoolConfig` | `confluent_flink_compute_pool_config` | settings | `GET /fcpm/v2/compute-pool-config` |
| `flink.confluent.hfl.dev/v1alpha1` `FlinkConnection` | `confluent_flink_connection` | crud | `GET /sql/v1/organizations/{organization_id}/environments/{environment_id}/connections/{connection_name}` |
| `flink.confluent.hfl.dev/v1alpha1` `FlinkMaterializedTable` | `confluent_flink_materialized_table` | crud | `GET /sql/v1/organizations/{organization_id}/environments/{environment_id}/databases/{kafka_cluster_id}/materialized-tables/{table_name}` |
| `flink.confluent.hfl.dev/v1alpha1` `FlinkStatement` | `confluent_flink_statement` | crud | `GET /sql/v1/organizations/{organization_id}/environments/{environment_id}/statements/{statement_name}` |
| `gateway.confluent.hfl.dev/v1alpha1` `Gateway` | `confluent_gateway` | crud | `GET /networking/v1/gateways/{id}` |
| `group.confluent.hfl.dev/v1alpha1` `GroupMapping` | `confluent_group_mapping` | crud | `GET /iam/v2/sso/group-mappings/{id}` |
| `identity.confluent.hfl.dev/v1alpha1` `IdentityPool` | `confluent_identity_pool` | crud | `GET /iam/v2/identity-providers/{provider_id}/identity-pools/{id}` |
| `identity.confluent.hfl.dev/v1alpha1` `IdentityProvider` | `confluent_identity_provider` | crud | `GET /iam/v2/identity-providers/{id}` |
| `invitation.confluent.hfl.dev/v1alpha1` `Invitation` | `confluent_invitation` | crud | `GET /iam/v2/invitations/{id}` |
| `ip.confluent.hfl.dev/v1alpha1` `IpFilter` | `confluent_ip_filter` | crud | `GET /iam/v2/ip-filters/{id}` |
| `ip.confluent.hfl.dev/v1alpha1` `IpGroup` | `confluent_ip_group` | crud | `GET /iam/v2/ip-groups/{id}` |
| `kafka.confluent.hfl.dev/v1alpha1` `KafkaAcl` | `confluent_kafka_acl` | assoc | `GET /kafka/v3/clusters/{cluster_id}/acls` |
| `kafka.confluent.hfl.dev/v1alpha1` `KafkaClientQuota` | `confluent_kafka_client_quota` | crud | `GET /kafka-quotas/v1/client-quotas/{id}` |
| `kafka.confluent.hfl.dev/v1alpha1` `KafkaCluster` | `confluent_kafka_cluster` | crud | `GET /cmk/v2/clusters/{id}` |
| `kafka.confluent.hfl.dev/v1alpha1` `KafkaClusterConfig` | `confluent_kafka_cluster_config` | upsert | `GET /kafka/v3/clusters/{cluster_id}/broker-configs/{name}` |
| `kafka.confluent.hfl.dev/v1alpha1` `KafkaMirrorTopic` | `confluent_kafka_mirror_topic` | crud | `GET /kafka/v3/clusters/{cluster_id}/links/{link_name}/mirrors/{mirror_topic_name}` |
| `ksql.confluent.hfl.dev/v1alpha1` `KsqlCluster` | `confluent_ksql_cluster` | crud | `GET /ksqldbcm/v2/clusters/{id}` |
| `network.confluent.hfl.dev/v1alpha1` `Network` | `confluent_network` | crud | `GET /networking/v1/networks/{id}` |
| `network.confluent.hfl.dev/v1alpha1` `NetworkLinkEndpoint` | `confluent_network_link_endpoint` | crud | `GET /networking/v1/network-link-endpoints/{id}` |
| `network.confluent.hfl.dev/v1alpha1` `NetworkLinkService` | `confluent_network_link_service` | crud | `GET /networking/v1/network-link-services/{id}` |
| `notifications.confluent.hfl.dev/v1alpha1` `NotificationsIntegration` | `confluent_notifications_integration` | crud | `GET /notifications/v1/integrations/{id}` |
| `notifications.confluent.hfl.dev/v1alpha1` `NotificationsResourcePreference` | `confluent_notifications_resource_preference` | crud | `GET /notifications/v1/resource-preferences/{id}` |
| `notifications.confluent.hfl.dev/v1alpha1` `NotificationsResourceSubscription` | `confluent_notifications_resource_subscription` | crud | `GET /notifications/v1/resource-subscriptions/{id}` |
| `notifications.confluent.hfl.dev/v1alpha1` `NotificationsSubscription` | `confluent_notifications_subscription` | crud | `GET /notifications/v1/subscriptions/{id}` |
| `peering.confluent.hfl.dev/v1alpha1` `Peering` | `confluent_peering` | crud | `GET /networking/v1/peerings/{id}` |
| `plugin.confluent.hfl.dev/v1alpha1` `Plugin` | `confluent_plugin` | crud | `GET /ccpm/v1/plugins/{id}` |
| `private.confluent.hfl.dev/v1alpha1` `PrivateLinkAccess` | `confluent_private_link_access` | crud | `GET /networking/v1/private-link-accesses/{id}` |
| `private.confluent.hfl.dev/v1alpha1` `PrivateLinkAttachment` | `confluent_private_link_attachment` | crud | `GET /networking/v1/private-link-attachments/{id}` |
| `private.confluent.hfl.dev/v1alpha1` `PrivateLinkAttachmentConnection` | `confluent_private_link_attachment_connection` | crud | `GET /networking/v1/private-link-attachment-connections/{id}` |
| `provider.confluent.hfl.dev/v1alpha1` `ProviderIntegration` | `confluent_provider_integration` | crud | `GET /pim/v1/integrations/{id}` |
| `provider.confluent.hfl.dev/v1alpha1` `ProviderIntegrationV2` | `confluent_provider_integration_authorization`, `confluent_provider_integration_setup` | crud | `GET /pim/v2/integrations/{id}` |
| `rtce.confluent.hfl.dev/v1alpha1` `RtceTopic` | `confluent_rtce_topic` | crud | `GET /rtce/v1/rtce-topics/{topic_name}` |
| `schema.confluent.hfl.dev/v1alpha1` `SchemaRegistryClusterConfig` | `confluent_schema_registry_cluster_config` | upsert | `GET /config` |
| `schema.confluent.hfl.dev/v1alpha1` `SchemaRegistryClusterMode` | `confluent_schema_registry_cluster_mode` | settings | `GET /mode` |
| `schema.confluent.hfl.dev/v1alpha1` `SchemaRegistryDek` | `confluent_schema_registry_dek` | crud | `GET /dek-registry/v1/keks/{name}/deks/{subject}` |
| `schema.confluent.hfl.dev/v1alpha1` `SchemaRegistryKek` | `confluent_schema_registry_kek` | crud | `GET /dek-registry/v1/keks/{name}` |
| `service.confluent.hfl.dev/v1alpha1` `ServiceAccount` | `confluent_service_account` | crud | `GET /iam/v2/service-accounts/{id}` |
| `subject.confluent.hfl.dev/v1alpha1` `SubjectConfig` | `confluent_subject_config` | upsert | `GET /config/{subject}` |
| `subject.confluent.hfl.dev/v1alpha1` `SubjectMode` | `confluent_subject_mode` | upsert | `GET /mode/{subject}` |
| `switchover.confluent.hfl.dev/v1alpha1` `SwitchoverEndpoint` | `confluent_switchover_endpoint` | crud | `GET /switchover/v1/switchover-endpoints/{id}` |
| `switchover.confluent.hfl.dev/v1alpha1` `SwitchoverPair` | `confluent_switchover_pair` | crud | `GET /switchover/v1/switchover-pairs/{id}` |
| `tableflow.confluent.hfl.dev/v1alpha1` `TableflowTopic` | `confluent_tableflow_topic` | crud | `GET /tableflow/v1/tableflow-topics/{display_name}` |
| `tag.confluent.hfl.dev/v1alpha1` `Tag` | `confluent_tag` | crud | `GET /catalog/v1/types/tagdefs/{tagName}` |
| `tag.confluent.hfl.dev/v1alpha1` `TagBinding` | `confluent_tag_binding` | crud | `GET /catalog/v1/entity/type/{typeName}/name/{qualifiedName}/tags` |
| `transit.confluent.hfl.dev/v1alpha1` `TransitGatewayAttachment` | `confluent_transit_gateway_attachment` | crud | `GET /networking/v1/transit-gateway-attachments/{id}` |

## Data-plane kinds

These kinds call an endpoint other than api.confluent.cloud, with that
endpoint's own API key. Make a second Connection (and Secret, with
`apiKey` and `apiSecret`) for each endpoint, and name it in each object's
`spec.connectionRef.name`:

| Kinds | Connection `baseURL` | Secret's `apiKey` / `apiSecret` |
| --- | --- | --- |
| `SchemaRegistryClusterConfig`, `SchemaRegistryClusterMode`, `SubjectConfig`, `SubjectMode`, `SchemaRegistryKek`, `SchemaRegistryDek` | the Schema Registry cluster's REST endpoint, `https://psrc-xxxxx.<region>.<cloud>.confluent.cloud` | a Schema Registry API key (an API key whose resource is the Schema Registry cluster) |
| `Tag`, `BusinessMetadata`, `TagBinding`, `BusinessMetadataBinding`, `CatalogEntityAttributes` (Stream Catalog, `/catalog/v1`) | the same Schema Registry endpoint | the same Schema Registry API key |
| `KafkaClusterConfig`, `KafkaAcl`, `KafkaMirrorTopic` | the Kafka cluster's REST endpoint, `https://pkc-xxxxx.<region>.<cloud>.confluent.cloud:443` | a Kafka API key for that cluster |
| `FlinkStatement`, `FlinkMaterializedTable`, `FlinkConnection` (Flink SQL, `/sql/v1`) | the region's Flink endpoint, `https://flink.<region>.<cloud>.confluent.cloud` | a Flink API key for the environment and region |

```yaml
apiVersion: core.confluent.hfl.dev/v1alpha1
kind: Connection
metadata:
  name: schema-registry
  namespace: team-a
spec:
  baseURL: https://psrc-xxxxx.us-east-2.aws.confluent.cloud
  secretRef:
    name: schema-registry-credentials   # apiKey, apiSecret: a Schema Registry API key
---
apiVersion: subject.confluent.hfl.dev/v1alpha1
kind: SubjectConfig
metadata:
  name: orders-value
  namespace: team-a
spec:
  connectionRef:
    name: schema-registry
  subject: orders-value
  compatibility: BACKWARD
```

Cluster-scoped paths carry the cluster in the URL: `KafkaClusterConfig`
and `KafkaAcl` take `spec.clusterID` (`lkc-...`) as well as the Connection for that
cluster's endpoint.

A few behave differently from their Terraform resources:

- `KafkaClusterConfig` is one broker setting per object (`spec.name`,
  `spec.value`); Terraform's `confluent_kafka_cluster_config` sets a map
  of them at once. Deleting the object resets the setting to its default.
- `SchemaRegistryClusterConfig` and `SubjectConfig` take `compatibility`
  and read it back as `compatibilityLevel`, so a change made outside
  Kubernetes to it is not seen as drift; it is sent whenever the spec
  changes. Deleting a `SchemaRegistryClusterConfig` resets the registry's
  config (`DELETE /config`), where Terraform leaves it as it is.
- `Tag` and `BusinessMetadata` cannot be updated in place (the API updates
  them with a PUT of a list on the collection); a changed spec is reported
  as drift.
- `KafkaAcl` is a link: made by `POST .../acls`, found in the cluster's
  list, and deleted by `DELETE .../acls` with its seven fields in the query
  (which deletes every ACL that matches them, as Terraform's delete does).
- `TagBinding` and `BusinessMetadataBinding` name their entity by
  `spec.entityType` and `spec.entityName` and their tag or business
  metadata by `spec.typeName`, the create body's own fields, which also
  fill the read's and delete's path. A `BusinessMetadataBinding`'s
  attributes change by a `PUT` of the list.
- `SwitchoverPair` and `SwitchoverEndpoint` take the environment's ID in
  `spec.environment` (sent as `?environment=` on every call but the create,
  never in the body), beside the CRNs the body carries.
- `ProviderIntegrationV2` stands in for both
  `confluent_provider_integration_setup` (its create) and
  `confluent_provider_integration_authorization` (an update of
  `spec.config`).

`KafkaMirrorTopic` is made on its cluster link and deleted as the topic it
is (`DELETE .../topics/{name}`, the name its read answers). Its
`mirror_status` is read, not managed: pausing, resuming, promoting or
failing it over (Terraform's `status`) is not done.

## Not covered

<!-- BEGIN generated: not covered (krmgen.py, from vendor.yaml hardBlocks and draftNotes) -->
1 of the 64 Terraform gap types have no kind: 1 is not built (`hardBlocks` in `hack/gen/vendor.yaml`).

| Terraform | Status | Why |
| --- | --- | --- |
| `confluent_tf_importer` | not built | Not an API resource: it reads existing resources and writes Terraform configuration and state for them, on the machine running Terraform. |
<!-- END generated: not covered -->

`FlinkConnection`'s credentials are `spec.spec.auth_data.data`, a JSON
string of the connection type's secret keys (`{"API_KEY": "..."}`), never
read back: keep it in a Secret and name it with
`secretRefs: {"spec.auth_data.data": {name, key}}`, beside
`spec.auth_data.kind: PlaintextProvider`. Only the credentials can be
changed, as Terraform's update allows.

`CatalogEntityAttributes` sets attributes of an entity the Stream Catalog
has, named by `spec.typeName` and `spec.attributes.qualifiedName`.
Deleting the object leaves them as they are; Terraform's destroy writes
each as `""`, which the spec can do first.

## For agents (MCP)

`cmd/mcp` serves this repository's kinds over the Model Context Protocol
(stdio), from `types.yaml`, `config/crds` and the live tests' records.
`.mcp.json` registers it, so an agent session opened in this repository
has it without setup. Elsewhere, run it from the repository root:

```sh
go run ./cmd/mcp
```

Its tools: `list_kinds`, `find_kinds_for_terraform`, `describe_kind` (every
field, the REST calls, the live record), `example` (the live-tested spec
where there is one), `validate_manifest` (the CRD schema, as the API server
checks it) and `plan_on_fake` (the controller run against the fake API,
with every REST call it makes). Nothing it does reaches the real API.
krm-rest-runtime's README ("For agents") has the details.

## Regenerate

```sh
hack/gen/generate.sh   # needs python3 + pyyaml, and krm-rest-runtime beside this repo
go test ./...
```
