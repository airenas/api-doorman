import http from "k6/http";
import { check, sleep } from "k6"; 

const admURL = 'http://host.docker.internal:8001';
const testURL = 'http://host.docker.internal:8000';
const expectedQuota = __ENV.EXPECTED_REQ * 10

export default function (data) {
    var url = testURL + '/private?key=' + data.key;
    var payload = JSON.stringify({
        text: '0123456789',
    });
    var params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    let res = http.post(url, payload, params);
    check(res, {
        "status was 200": (r) => r.status == 200,
        "transaction time OK": (r) => r.timings.duration < 200
    });
    sleep(0.1);
}

export function setup() {
    var url = admURL + '/key';
    var payload = JSON.stringify({
        limit: expectedQuota * 2,
        validTo: '2030-11-24T11:07:00Z'
    });
    var params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    let res = http.post(url, payload, params);
    console.log("Test key: " + res.json().key);
    return { key: res.json().key };
}

export function teardown(data) {
    var url = admURL + '/key/' + data.key;
    console.log("Url: " + url);
    let res = http.get(url);
    let jRes = res.json().key
    console.log("Final quota: " + jRes.quotaValue + " expected: " + expectedQuota);
    check(res, {
        "quota": (r) => r.json().key.quotaValue == expectedQuota,
    });
}