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
  trend,
  trendUp,
}: StatCardProps) {
  return (
    <div className="enterprise-card rounded-lg p-5 hover-lift group">
      <div className="flex items-start justify-between mb-3">
        <div className="min-w-0 flex-1">
          <p className="text-[11px] font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-1">
            {title}
          </p>
          <p className="text-2xl font-bold text-slate-800 dark:text-white tracking-tight truncate">
            {value}
          </p>
          {subtitle && (
            <p className="text-xs text-slate-400 mt-0.5 truncate">{subtitle}</p>
          )}
        </div>
        {/* Icon — small, flat, muted */}
        <div className="w-9 h-9 rounded-md bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-base flex-shrink-0 ml-3">
          {icon}
        </div>
      </div>

      {trend && (
        <div className="flex items-center gap-1.5 pt-3 border-t border-slate-100 dark:border-slate-800">
          <span
            className={`text-xs font-semibold ${
              trendUp === undefined
                ? 'text-slate-400'
                : trendUp
                ? 'text-emerald-600 dark:text-emerald-400'
                : 'text-red-600 dark:text-red-400'
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
