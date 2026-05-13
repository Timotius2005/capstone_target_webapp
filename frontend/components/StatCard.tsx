interface StatCardProps {
  title: string
  value: string
  subtitle?: string
  icon: string
  iconBg: string
  trend?: string
  trendUp?: boolean
}

export default function StatCard({
  title,
  value,
  subtitle,
  icon,
  iconBg,
  trend,
  trendUp,
}: StatCardProps) {
  return (
    <div className="glass-card rounded-2xl p-6 hover-lift group">
      <div className="flex items-start justify-between mb-4">
        <div className="min-w-0 flex-1">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
            {title}
          </p>
          <p className="text-2xl font-bold text-slate-800 dark:text-white tracking-tight truncate">
            {value}
          </p>
          {subtitle && (
            <p className="text-xs text-slate-400 mt-1 truncate">{subtitle}</p>
          )}
        </div>
        <div
          className={`w-12 h-12 rounded-xl flex items-center justify-center text-xl shadow-lg flex-shrink-0 ml-3 ${iconBg}`}
        >
          {icon}
        </div>
      </div>

      {trend && (
        <div className="flex items-center gap-1.5 pt-3 border-t border-slate-200/30 dark:border-slate-700/30">
          <span
            className={`text-xs font-semibold ${
              trendUp === undefined
                ? 'text-slate-400'
                : trendUp
                ? 'text-emerald-400'
                : 'text-red-400'
            }`}
          >
            {trendUp === undefined ? '→' : trendUp ? '↑' : '↓'} {trend}
          </span>
          <span className="text-xs text-slate-400">vs bulan lalu</span>
        </div>
      )}
    </div>
  )
}
