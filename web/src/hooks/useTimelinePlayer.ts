import { useState, useRef, useCallback, useEffect } from 'react';

export interface TimelinePlayer {
  currentSecond: number;
  isPlaying: boolean;
  playbackSpeed: number;
  totalSeconds: number;
  play: () => void;
  pause: () => void;
  toggle: () => void;
  seek: (second: number) => void;
  setSpeed: (speed: number) => void;
  isFinished: boolean;
}

export function useTimelinePlayer(totalSeconds: number): TimelinePlayer {
  const [currentSecond, setCurrentSecond] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const [isFinished, setIsFinished] = useState(false);
  const lastTimeRef = useRef<number>(0);
  const accumulatorRef = useRef<number>(0);
  const rafRef = useRef<number>(0);

  const tick = useCallback((timestamp: number) => {
    if (lastTimeRef.current === 0) {
      lastTimeRef.current = timestamp;
    }

    const delta = (timestamp - lastTimeRef.current) / 1000;
    lastTimeRef.current = timestamp;
    accumulatorRef.current += delta * playbackSpeed;

    if (accumulatorRef.current >= 1) {
      const steps = Math.floor(accumulatorRef.current);
      accumulatorRef.current -= steps;

      setCurrentSecond((prev) => {
        const next = prev + steps;
        if (next >= totalSeconds - 1) {
          setIsPlaying(false);
          setIsFinished(true);
          return totalSeconds - 1;
        }
        return next;
      });
    }

    rafRef.current = requestAnimationFrame(tick);
  }, [playbackSpeed, totalSeconds]);

  useEffect(() => {
    if (isPlaying) {
      lastTimeRef.current = 0;
      accumulatorRef.current = 0;
      rafRef.current = requestAnimationFrame(tick);
    }
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [isPlaying, tick]);

  const play = useCallback(() => {
    if (currentSecond >= totalSeconds - 1) {
      setCurrentSecond(0);
      setIsFinished(false);
    }
    setIsPlaying(true);
  }, [currentSecond, totalSeconds]);

  const pause = useCallback(() => setIsPlaying(false), []);
  const toggle = useCallback(() => {
    if (isPlaying) pause();
    else play();
  }, [isPlaying, play, pause]);

  const seek = useCallback((second: number) => {
    setCurrentSecond(Math.max(0, Math.min(second, totalSeconds - 1)));
    setIsFinished(second >= totalSeconds - 1);
  }, [totalSeconds]);

  const setSpeed = useCallback((speed: number) => {
    setPlaybackSpeed(speed);
  }, []);

  return {
    currentSecond,
    isPlaying,
    playbackSpeed,
    totalSeconds,
    play,
    pause,
    toggle,
    seek,
    setSpeed,
    isFinished,
  };
}
