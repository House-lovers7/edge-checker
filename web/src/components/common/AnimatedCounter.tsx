import { motion, useSpring, useTransform } from 'framer-motion';
import { useEffect } from 'react';

interface Props {
  value: number;
  format?: (v: number) => string;
  className?: string;
}

export function AnimatedCounter({ value, format, className = '' }: Props) {
  const spring = useSpring(0, { stiffness: 80, damping: 20 });
  const display = useTransform(spring, (v) => {
    const rounded = Math.round(v);
    return format ? format(rounded) : rounded.toLocaleString();
  });

  useEffect(() => {
    spring.set(value);
  }, [value, spring]);

  return <motion.span className={className}>{display}</motion.span>;
}
