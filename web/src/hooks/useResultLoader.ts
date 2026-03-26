import { useState, useCallback } from 'react';
import type { Result } from '../types/result';

export function useResultLoader() {
  const [result, setResult] = useState<Result | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadFromFile = useCallback((file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = JSON.parse(e.target?.result as string) as Result;
        if (!data.scenario_name || !data.timeline) {
          setError('Invalid result JSON: missing scenario_name or timeline');
          return;
        }
        setError(null);
        setResult(data);
      } catch {
        setError('Failed to parse JSON file');
      }
    };
    reader.readAsText(file);
  }, []);

  const loadResult = useCallback((data: Result) => {
    setError(null);
    setResult(data);
  }, []);

  const reset = useCallback(() => {
    setResult(null);
    setError(null);
  }, []);

  return { result, error, loadFromFile, loadResult, reset };
}
