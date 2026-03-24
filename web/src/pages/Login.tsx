import { useState } from "react";
import { useAuthStore } from "../stores/authStore";
import { Shield, Eye, EyeOff } from "lucide-react";

export function Login() {
  const { setup, login, setup_password } = useAuthStore();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [showPw, setShowPw] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (!setup && password !== confirm) {
      setError("Passwords do not match");
      return;
    }
    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    setLoading(true);
    try {
      if (setup) {
        await login(password);
      } else {
        await setup_password(password);
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Authentication failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{
      height: "100vh",
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      backgroundColor: "#0f1117",
    }}>
      <div style={{
        width: "360px",
        backgroundColor: "#161b27",
        border: "1px solid #2a2e3a",
        borderRadius: "12px",
        padding: "40px",
      }}>
        <div style={{ textAlign: "center", marginBottom: "32px" }}>
          <div style={{
            width: "48px",
            height: "48px",
            borderRadius: "12px",
            backgroundColor: "#1e3a5f",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            margin: "0 auto 16px",
          }}>
            <Shield size={24} color="#4a9eff" />
          </div>
          <h1 style={{ color: "#e2e8f0", fontSize: "20px", fontWeight: 600, margin: 0 }}>
            {setup ? "Welcome back" : "Set up NetMap"}
          </h1>
          <p style={{ color: "#64748b", fontSize: "14px", marginTop: "8px" }}>
            {setup ? "Enter your password to continue" : "Create a password to secure your instance"}
          </p>
        </div>

        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: "16px" }}>
            <label style={{ display: "block", color: "#94a3b8", fontSize: "13px", marginBottom: "6px" }}>
              Password
            </label>
            <div style={{ position: "relative" }}>
              <input
                type={showPw ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password"
                autoFocus
                style={{
                  width: "100%",
                  padding: "10px 40px 10px 12px",
                  backgroundColor: "#1a2035",
                  border: "1px solid #2a2e3a",
                  borderRadius: "8px",
                  color: "#e2e8f0",
                  fontSize: "14px",
                  outline: "none",
                  boxSizing: "border-box",
                }}
              />
              <button
                type="button"
                onClick={() => setShowPw(!showPw)}
                style={{
                  position: "absolute",
                  right: "10px",
                  top: "50%",
                  transform: "translateY(-50%)",
                  background: "none",
                  border: "none",
                  cursor: "pointer",
                  color: "#64748b",
                  padding: 0,
                }}
              >
                {showPw ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
          </div>

          {!setup && (
            <div style={{ marginBottom: "16px" }}>
              <label style={{ display: "block", color: "#94a3b8", fontSize: "13px", marginBottom: "6px" }}>
                Confirm Password
              </label>
              <input
                type={showPw ? "text" : "password"}
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                placeholder="Confirm password"
                style={{
                  width: "100%",
                  padding: "10px 12px",
                  backgroundColor: "#1a2035",
                  border: "1px solid #2a2e3a",
                  borderRadius: "8px",
                  color: "#e2e8f0",
                  fontSize: "14px",
                  outline: "none",
                  boxSizing: "border-box",
                }}
              />
            </div>
          )}

          {error && (
            <div style={{
              padding: "10px 12px",
              backgroundColor: "#2d1515",
              border: "1px solid #7f1d1d",
              borderRadius: "8px",
              color: "#fca5a5",
              fontSize: "13px",
              marginBottom: "16px",
            }}>
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !password}
            style={{
              width: "100%",
              padding: "10px",
              backgroundColor: loading || !password ? "#1e3a5f" : "#2563eb",
              color: loading || !password ? "#64748b" : "#fff",
              border: "none",
              borderRadius: "8px",
              fontSize: "14px",
              fontWeight: 600,
              cursor: loading || !password ? "not-allowed" : "pointer",
            }}
          >
            {loading ? "Please wait…" : setup ? "Sign In" : "Set Password & Continue"}
          </button>
        </form>
      </div>
    </div>
  );
}
