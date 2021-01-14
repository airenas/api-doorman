# DB Schema

## Key table

Table keeps keys

| Field | Type | Usage |
| ---|-|-|
| key[*pk*] | string | Key or IP if created automatically |
| manual[pk*] | bool | *false* - indicates IP as a key, *true* - manually created key |
||
| limit   | float64 | Quota limit value - total value granted to the key |
| quotaValue | float64 | Value of current quota usage (in total) |
| quotaValueFailed | float64 | Value of quota requested, but failed because of the limit or other errors (in total) |
||
| validTo | time | Disables key after this time |
| disabled | bool | Indicates if the key is disabled |
||
| created | time | Key creation time |
| updated | time | Key update time |
| lastUsed | time | Time of the last access |
| lastIP | string | IP of the last access |

*pk* - primary key

---

## Log table

Table keeps all requests. System logs user`s IP, time, quota value of the request.

| Laukas| Tipas | Paskirtis |
| ---|-|-|
| key | string | |
| url | string | request URL, path |
| quotaValue | float64 | Quota used by the request |
| date | time | Time of the request |
| ip | string | IP of the user |
| fail | bool | *true* if the requests has failed |
| response | int | Code of the response returned to the user |
