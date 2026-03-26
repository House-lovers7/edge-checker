import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts';
import { getStatusColor } from '../../utils/colors';

interface Props {
  statusCounts: Record<string, number>;
}

export function StatusDonut({ statusCounts }: Props) {
  const data = Object.entries(statusCounts)
    .filter(([, count]) => count > 0)
    .map(([code, count]) => ({
      name: code,
      value: count,
      color: getStatusColor(code),
    }))
    .sort((a, b) => b.value - a.value);

  const total = data.reduce((sum, d) => sum + d.value, 0);

  if (total === 0) {
    return (
      <div className="p-4 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-center">
        <div className="text-xs text-[var(--color-text-muted)] mb-2">Status Distribution</div>
        <div className="text-[var(--color-text-muted)] text-sm py-8">Waiting for data...</div>
      </div>
    );
  }

  return (
    <div className="p-4 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)]">
      <div className="text-xs text-[var(--color-text-muted)] mb-2">Status Distribution</div>
      <div className="h-48 relative">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              innerRadius={50}
              outerRadius={75}
              paddingAngle={2}
              dataKey="value"
              isAnimationActive={true}
              animationDuration={500}
            >
              {data.map((entry) => (
                <Cell key={entry.name} fill={entry.color} />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="text-center">
            <div className="text-xl font-bold">{total.toLocaleString()}</div>
            <div className="text-[10px] text-[var(--color-text-muted)]">total</div>
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-3 mt-2 justify-center">
        {data.map((d) => (
          <div key={d.name} className="flex items-center gap-1 text-xs">
            <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: d.color }} />
            <span className="text-[var(--color-text-muted)]">{d.name}:</span>
            <span>{d.value.toLocaleString()}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
