import React from "react";
import { useMsal } from "@azure/msal-react";
import { loginRequest } from "@/authConfig";
import { toast } from "sonner";

const Login = () => {
  const { instance } = useMsal();

  const login = () => {
    instance
      .loginPopup(loginRequest)
      .then(() => {
        toast.success("Login successful!");
      })
      .catch((error) => {
        toast.error("Login failed!");
      });
  };

  return (
    <div className="h-[100dvh] w-full bg-gradient-to-br from-gray-950 to-gray-800 flex items-center justify-center p-4 sm:p-6 lg:p-8">
      <div
        className="absolute inset-0 opacity-10 pointer-events-none"
        style={{
          backgroundImage: "radial-gradient(ellipse at center, oklch(0.9247 0.0524 66.1732 / 0.15) 0%, transparent 70%)",
          backgroundSize: "200% 200%",
          animation: "pulse 10s infinite alternate",
        }}
      ></div>
      <div className="relative z-10 bg-white/5 backdrop-blur-md border border-white/10 rounded-3xl shadow-xl overflow-hidden transform transition-all duration-300 ease-in-out hover:scale-[1.01] w-full max-w-sm mx-auto p-6 sm:p-8">
        <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/10 to-purple-500/10 rounded-3xl pointer-events-none"></div>
        <div className="relative z-10 flex flex-col items-center justify-center">
          <h3 className="text-4xl font-extrabold text-white text-center mb-6 drop-shadow-lg tracking-tight">Welcome Back!</h3>
          <p className="text-gray-300 text-lg mb-8 text-center max-w-xs">Securely access your account to continue.</p>

          <button
            onClick={login}
            className="relative w-full py-3 px-6 bg-white text-gray-800 rounded-lg shadow-md hover:shadow-lg transition-all duration-300 ease-in-out transform hover:-translate-y-0.5 flex items-center justify-center gap-3 text-lg font-semibold border border-gray-300/50 group"
          >
            <img src="./Microsoft_Logo.svg" alt="Microsoft Logo" className="w-7 h-7 transition-all duration-300" />
            <span>Sign in with Microsoft</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default Login;
