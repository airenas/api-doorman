// test /quotatts/private endpoint
// with expected doscount 0.8
import http from "k6/http";
import { check, sleep } from "k6";
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const prj = "test"
const admURL = 'http://host.docker.internal:8001';
const testURL = 'http://host.docker.internal:8000';
const expectedQuota = __ENV.EXPECTED_REQ * 10 * 0.8;

export let options = {
    thresholds: {
        checks: ['rate==1'],
    },
};

export default function (data) {
    var url = testURL + '/quotatts/private';
    var payload = JSON.stringify({
        text: '0123456789',
        saveRequest: true
    });
    var params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Key ' + data.key
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
    const id = uuidv4();
    var payload = JSON.stringify({
        validTo: '2034-11-24T11:07:00Z',
        service: prj,
        credits: expectedQuota * 2,
        id: id,
    });
    var params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Key olia'
        },
    };
    let res = http.post(url, payload, params);
    console.log("Response body: ", res.body);
    console.log("Test key: " + res.json().key);
    return { key: res.json().key, id: id };
}

export function teardown(data) {
    var url = admURL + '/' + prj + '/key/' + data.id;

    var params = {
        headers: {
            'Authorization': 'Key olia'
        },
    };
    let res = http.get(url, params);

    let jRes = res.json().key;
    let qv = jRes.quotaValue;
    console.log("Final quota: " + qv + " expected: " + expectedQuota);
    console.log("Total quota: " + jRes.limit);
    check(res, {
        "quota": (r) => qv == expectedQuota,
    });
}