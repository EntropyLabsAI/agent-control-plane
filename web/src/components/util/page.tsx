import React from "react";

export default function Page({ children, title, subtitle, icon, cols = 1 }: { children: React.ReactNode, title: string, subtitle?: React.ReactNode, icon: React.ReactNode, cols?: number }) {

  return (
    <div className="px-32 pt-32 flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <div className="flex flex-row gap-4 items-center">
          {icon}
          <h1 className="text-2xl font-bold">{title}</h1>
        </div>
        {subtitle && <p className="text-sm text-gray-500">{subtitle}</p>}

      </div>
      <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-${cols} gap-6`}>
        {children}
      </div>
    </div>
  )
}
