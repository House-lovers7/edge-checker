import { useCallback, useState } from 'react';
import { motion } from 'framer-motion';
import type { Result } from '../../types/result';
import { sampleResult } from '../../utils/sampleData';

interface Props {
  onLoad: (result: Result) => void;
  onFileLoad: (file: File) => void;
  error: string | null;
}

export function FileDropZone({ onLoad, onFileLoad, error }: Props) {
  const [isDragging, setIsDragging] = useState(false);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files[0];
    if (file && file.name.endsWith('.json')) {
      onFileLoad(file);
    }
  }, [onFileLoad]);

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) onFileLoad(file);
  }, [onFileLoad]);

  return (
    <div className="min-h-screen flex items-center justify-center p-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ type: 'spring', stiffness: 100, damping: 15 }}
        className="max-w-lg w-full"
      >
        <div className="text-center mb-8">
          <motion.h1
            className="text-4xl font-bold mb-2"
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ type: 'spring', stiffness: 120, delay: 0.1 }}
          >
            edge-checker
          </motion.h1>
          <p className="text-[var(--color-text-muted)]">
            WAF/CDN Defense Verification Visualizer
          </p>
        </div>

        <motion.div
          onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
          onDragLeave={() => setIsDragging(false)}
          onDrop={handleDrop}
          animate={{
            borderColor: isDragging ? '#3b82f6' : '#334155',
            backgroundColor: isDragging ? 'rgba(59,130,246,0.05)' : 'transparent',
          }}
          className="border-2 border-dashed rounded-xl p-12 text-center cursor-pointer transition-colors"
          onClick={() => document.getElementById('file-input')?.click()}
        >
          <div className="text-5xl mb-4">
            {isDragging ? '📥' : '📄'}
          </div>
          <p className="text-lg mb-2">
            Drop result JSON here
          </p>
          <p className="text-sm text-[var(--color-text-muted)]">
            or click to select file
          </p>
          <input
            id="file-input"
            type="file"
            accept=".json"
            onChange={handleFileSelect}
            className="hidden"
          />
        </motion.div>

        {error && (
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-[var(--color-blocked)] text-sm mt-4 text-center"
          >
            {error}
          </motion.p>
        )}

        <div className="mt-6 text-center">
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={() => onLoad(sampleResult)}
            className="px-6 py-3 bg-[var(--color-accent)] text-white rounded-lg font-medium hover:opacity-90 transition-opacity"
          >
            Try with sample data
          </motion.button>
        </div>
      </motion.div>
    </div>
  );
}
