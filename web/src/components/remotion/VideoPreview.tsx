import { Player } from '@remotion/player';
import { ResultComposition } from './ResultComposition';
import type { Result } from '../../types/result';
import { motion } from 'framer-motion';

interface Props {
  result: Result;
  onClose: () => void;
}

const FPS = 30;

export function VideoPreview({ result, onClose }: Props) {
  const totalSeconds = result.timeline.length;
  const durationInFrames = totalSeconds * FPS + FPS; // +1s for verdict display

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-8"
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div className="w-full max-w-4xl">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-bold">Video Preview</h3>
          <button onClick={onClose} className="text-[var(--color-text-muted)] hover:text-white text-2xl">&times;</button>
        </div>
        <div className="rounded-xl overflow-hidden border border-[var(--color-border)]">
          <Player
            component={ResultComposition}
            inputProps={{ result }}
            durationInFrames={durationInFrames}
            fps={FPS}
            compositionWidth={1280}
            compositionHeight={720}
            style={{ width: '100%' }}
            controls
            autoPlay={false}
          />
        </div>
        <p className="text-xs text-[var(--color-text-muted)] mt-3 text-center">
          Use Remotion CLI to export as MP4: npx remotion render
        </p>
      </div>
    </motion.div>
  );
}
