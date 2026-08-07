/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { useCallback, useEffect, useRef, useState } from 'react';

interface FetchState<T> {
  data: T | undefined;
  loading: boolean;
  error: unknown;
  refresh: () => void;
}

export function useFetch<T>(
  fetcher: () => Promise<T>,
  deps: unknown[],
): FetchState<T> {
  const [data, setData] = useState<T | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(undefined);
  const [tick, setTick] = useState(0);

  const refresh = useCallback(() => {
    setTick(t => t + 1);
  }, []);

  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetcherRef
      .current()
      .then(result => {
        if (!cancelled) {
          setData(result);
          setError(undefined);
        }
      })
      .catch(err => {
        if (!cancelled) {
          setError(err);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, tick]);

  return { data, loading, error, refresh };
}

export function useMutation<Args extends unknown[], Result>(
  fn: (...args: Args) => Promise<Result>,
): {
  run: (...args: Args) => Promise<Result | undefined>;
  loading: boolean;
} {
  const [loading, setLoading] = useState(false);
  const fnRef = useRef(fn);
  fnRef.current = fn;

  const run = useCallback(async (...args: Args) => {
    setLoading(true);
    try {
      return await fnRef.current(...args);
    } finally {
      setLoading(false);
    }
  }, []);

  return { run, loading };
}
