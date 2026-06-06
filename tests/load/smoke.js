import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const errors = new Counter('errors');
const BASE_URLS = ['http://localhost:8080', 'http://localhost:8081'];

export const options = {
  vus: 5,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<<0.01'],
    http_req_duration: ['p99<<100'],
    errors: ['count<<1'],
  },
};

export default function () {
  const base = BASE_URLS[__VU % 2];
  const res = http.get(`${base}/health`);
  const ok = check(res, {
    'status 200': (r) => r.status === 200,
    'body has status': (r) => JSON.parse(r.body).status === 'ok',
  });
  if (!ok) errors.add(1);
}
