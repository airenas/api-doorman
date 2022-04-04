# DB Schema

## Key table

Table keeps keys

| Field | Type | Usage |
| ---|-|-|
| key[*pk**] | string | Key or IP if created automatically |
| manual[*pk*] | bool | *false* - indicates IP as a key, *true* - manually created key |
| IPWhiteList | string | Comma separated IP ranges in CIDR format. Eg.: *192.168.1.1/32,21.21.21.0/24* |
| description | string | Key description |
||
| limit   | float64 | Quota limit value - total value granted to the key |
| quotaValue | float64 | Value of current quota usage (in total) |
| quotaValueFailed | float64 | Value of quota requested, but failed because of the limit or other errors (in total) |
||
| validTo | time | Disables key after this time |
| disabled | bool | Indicates if the key is disabled |
||
| tags | []string | Array of tag values passed as headers to proxy. Value must contain *':'*. Sample: *"x-header : value"*. |
||
| created | time | Key creation time |
| updated[*ind**] | time | Key update time |
| lastUsed | time | Time of the last access |
| lastIP | string | IP of the last access |
| externalID | string | ID from external system |


*pk* - primary key

*ind* - index

---

## Log table

Table keeps all requests. System logs user`s IP, time, quota value of the request.

| Field | Type | Usage |
| ---|-|-|
| key[*ind*] | string | |
| url | string | Request URL, path |
| quotaValue | float64 | Quota used by the request |
| date[*ind*] | time | Time of the request |
| ip | string | IP of the user |
| fail | bool | *true* if the requests has failed |
| responseCode | int | Code of the response returned to the user |
| requestID[*ind*] | string | Generated unique ID for the request |
| errorMsg | string | Error msg of failed requuest |


## KeyMap table

Table maps keys with external IDs.

| Field | Type | Usage |
| ---|-|-|
| externalID[*pk*] | string | ID of external system |
| key[*ind*] | string | Current key |
| project | string | Name of a service. For example: *tts*, ... |
| created | time | Time of record creation |
| old | [] **oldKey** | Array of old keys for the externalID |

### OldKey structure

Strukture keeps obsolete keys.

| Field | Type | Usage |
| ---|-|-|
| key | string | Old key value|
| changedOn | time | Time of the key change |

## Operation table

Table keeps quota increase operations.

| Field | Type | Usage |
| ---|-|-|
| operationID[*pk*] | string | Unique ID of the operation |
| key | string | Key value |
| date | time | Time of operation |
| quotaValue | float64 | Quota increase value |
