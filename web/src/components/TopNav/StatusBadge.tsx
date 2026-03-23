import { motion } from "framer-motion";

interface StatusBadgeProps {
  count: number;
  label: string;
  color: "teal" | "red" | "zinc";
}

const colorMap = {
  teal: { bg: "bg-[#2dd4bf]/10", text: "text-[#2dd4bf]", dot: "bg-[#2dd4bf]" },
  red: { bg: "bg-[#ef4444]/10", text: "text-[#ef4444]", dot: "bg-[#ef4444]" },
  zinc: { bg: "bg-zinc-700/30", text: "text-zinc-400", dot: "bg-zinc-500" },
};

export function StatusBadge({ count, label, color }: StatusBadgeProps) {
  const c = colorMap[color];
  return (
    <motion.div
      className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md ${c.bg}`}
      initial={{ scale: 0.9, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      transition={{ type: "spring", stiffness: 500, damping: 30 }}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${c.dot}`} />
      <span className={`font-medium ${c.text}`}>{count}</span>
      <span className="text-zinc-500 text-xs">{label}</span>
    </motion.div>
  );
}
