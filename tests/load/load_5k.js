import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const allowedCounter = new Counter('requests_allowed');

export const options = {
  scenarios: {
    load: {
      executor: 'constant-arrival-rate',
      rate: 5000,
      timeUnit: '1s',
      duration: '60s',
      preAllocatedVUs: 200,
      maxVUs: 500,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: [
      'p(50)<5',
      'p(95)<15',
      'p(99)<30',
    ],
  },
};

const NODES = [
  'http://localhost:8080',
  'http://localhost:8081',
];

const PAYLOAD = JSON.stringify({
  client_id: 'bench-client-5k',
  algorithm: 'fixed_window',
  limit: 9999999,
  window_secs: 3600,
});

export default function () {
  const node = NODES[__VU % 2];

  const res = http.post(`${node}/check`, PAYLOAD, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'status 200': (r) => r.status === 200,
  });

  if (res.status === 200 && res.json().allowed) {
    allowedCounter.add(1);
  }
}

export function handleSummary(data) {
  const dur = data.metrics.http_req_duration.values;

  console.log('\n=== 5K RPS BENCHMARK RESULTS ===');
  console.log(`p50: ${dur.med.toFixed(2)}ms`);
  console.log(`p95: ${dur['p(95)'].toFixed(2)}ms`);
  console.log(`p90: ${dur['p(90)'].toFixed(2)}ms`);
  console.log(`avg: ${dur.avg.toFixed(2)}ms`);
  console.log(`max: ${dur.max.toFixed(2)}ms`);
  console.log(`reqs: ${data.metrics.http_reqs.values.count}`);
  console.log(
    `errs: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(3)}%`
  );

  return {};
}