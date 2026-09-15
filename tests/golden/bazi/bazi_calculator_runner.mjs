#!/usr/bin/env node

import { createInterface } from 'node:readline';
import { pathToFileURL } from 'node:url';
import { join } from 'node:path';

const root = process.env.BAZI_CALCULATOR_DIR;
if (!root) {
  process.stderr.write('BAZI_CALCULATOR_DIR must point to shenshuge/bazi-calculator\n');
  process.exit(2);
}

try {
  const calculator = await import(pathToFileURL(join(root, 'dist', 'index.js')));
  var computeBazi = calculator.computeBazi;
} catch (error) {
  process.stderr.write(`cannot load bazi-calculator from ${root}: ${error.message}\n`);
  process.exit(2);
}

const input = createInterface({ input: process.stdin });
input.on('line', line => {
  if (!line.trim()) {
    return;
  }
  const result = computeBazi(JSON.parse(line));
  const pillars = [
    result.year.stem + result.year.branch,
    result.month.stem + result.month.branch,
    result.day.stem + result.day.branch,
    result.hour.stem + result.hour.branch,
  ];
  process.stdout.write(`${JSON.stringify(pillars)}\n`);
});
