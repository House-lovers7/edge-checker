interface Props {
  x1: number;
  x2: number;
  y: number;
}

export function ConnectionPath({ x1, x2, y }: Props) {
  return (
    <line
      x1={x1 + 40}
      y1={y}
      x2={x2 - 40}
      y2={y}
      stroke="var(--color-border)"
      strokeWidth={2}
      strokeDasharray="8 4"
      opacity={0.5}
    />
  );
}
