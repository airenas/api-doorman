# DB Schema

## Key table

Table keeps keys

| Field | Type | Usage |
| ---|-|-|
| key[*pk*] | string | Key or IP if created automatically |
| manual[pk*] | bool | *false* - indicates IP as a key, *true* - manually created key |
| validTo | time | disables key after this time |
| limit   | float64 | Quata limit value |
| quotaValue | float64 | Current quota usage value |
| quotaValueFailed | float64 | Quota value requested, but failed because of the limit (in total) |
| created | time | Key creation time |
| updated | time | Key update time |
| lastUsed | time | Time of last access |
| lastIP | string | IP of last access |
| disabled | bool | Disables key |

*pk* - primary key

---

## Log table

Table keeps all requests. System logs users IP, time, quota value.

| Laukas| Tipas | Paskirtis |
| ---|-|-|
| key | string | |
| utl | string | request URL, path |
| quotaValue | float64 | Quota used by the request |
| date | time | Time of the request |
| ip | string | IP of the user |
| fail | bool | *true* if the requests has failed |
| response | int | Code of the response provided to user |
