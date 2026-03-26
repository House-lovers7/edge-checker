import { motion } from 'framer-motion';
import type { TimelinePlayer } from '../../hooks/useTimelinePlayer';

interface Props {
  player: TimelinePlayer;
}

const SPEEDS = [1, 2, 4];

export function PlaybackControls({ player }: Props) {
  return (
    <div className="flex items-center gap-4 px-6 py-3 bg-[var(--color-surface)]">
      {/* Play/Pause */}
      <motion.button
        whileTap={{ scale: 0.9 }}
        onClick={player.toggle}
        className="w-10 h-10 flex items-center justify-center rounded-full bg-[var(--color-accent)] text-white"
      >
        {player.isPlaying ? '⏸' : '▶'}
      </motion.button>

      {/* Seek bar */}
      <div className="flex-1 flex items-center gap-3">
        <span className="text-xs text-[var(--color-text-muted)] w-8 text-right">
          {player.currentSecond}s
        </span>
        <input
          type="range"
          min={0}
          max={player.totalSeconds - 1}
          value={player.currentSecond}
          onChange={(e) => player.seek(parseInt(e.target.value))}
          className="flex-1 h-1.5 rounded-full appearance-none bg-[var(--color-border)] accent-[var(--color-accent)] cursor-pointer"
        />
        <span className="text-xs text-[var(--color-text-muted)] w-8">
          {player.totalSeconds - 1}s
        </span>
      </div>

      {/* Speed */}
      <div className="flex items-center gap-1">
        {SPEEDS.map((speed) => (
          <button
            key={speed}
            onClick={() => player.setSpeed(speed)}
            className={`px-2 py-1 text-xs rounded ${
              player.playbackSpeed === speed
                ? 'bg-[var(--color-accent)] text-white'
                : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'
            }`}
          >
            {speed}x
          </button>
        ))}
      </div>
    </div>
  );
}
