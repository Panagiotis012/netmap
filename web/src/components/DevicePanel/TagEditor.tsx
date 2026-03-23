import { X, Plus } from "lucide-react";
import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

interface Props {
  tags: string[];
  onChange: (tags: string[]) => void;
}

export function TagEditor({ tags, onChange }: Props) {
  const [adding, setAdding] = useState(false);
  const [input, setInput] = useState("");

  const add = () => {
    const tag = input.trim();
    if (tag && !tags.includes(tag)) onChange([...tags, tag]);
    setInput("");
    setAdding(false);
  };

  const remove = (tag: string) => onChange(tags.filter((t) => t !== tag));

  return (
    <div style={{ display: "flex", flexWrap: "wrap", gap: "6px" }}>
      <AnimatePresence>
        {tags.map((tag) => (
          <motion.span
            key={tag}
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.8, opacity: 0 }}
            style={{ display: "flex", alignItems: "center", gap: "4px", background: "rgba(139,92,246,0.15)", color: "#a78bfa", fontSize: "12px", padding: "2px 8px", borderRadius: "100px" }}
          >
            {tag}
            <button onClick={() => remove(tag)} style={{ background: "none", border: "none", cursor: "pointer", color: "inherit", padding: 0 }}>
              <X size={10} />
            </button>
          </motion.span>
        ))}
      </AnimatePresence>
      {adding ? (
        <input
          autoFocus
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onBlur={add}
          onKeyDown={(e) => e.key === "Enter" && add()}
          style={{ fontSize: "12px", padding: "2px 8px", borderRadius: "100px", border: "1px solid #2a2e3a", background: "#0f1117", color: "#e4e4e7", outline: "none", width: "80px" }}
        />
      ) : (
        <button
          onClick={() => setAdding(true)}
          style={{ background: "none", border: "none", cursor: "pointer", color: "#71717a", padding: 0 }}
        >
          <Plus size={14} />
        </button>
      )}
    </div>
  );
}
