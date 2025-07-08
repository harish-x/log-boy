import { useMsal } from "@azure/msal-react";
import { useEffect } from "react";
import { loginRequest } from "@/authConfig";

const ProtectedRoute = ({ children }) => {
  const { instance, accounts, inProgress } = useMsal();

  useEffect(() => {
    if (accounts.length === 0 && inProgress === "none") {
      instance.loginRedirect(loginRequest);
    }
  }, [accounts, inProgress, instance]);

  if (inProgress !== "none") {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <span className="loader"></span>
      </div>
    );
  }

  if (accounts.length > 0) {
    return children;
  }

  return null;
};

export default ProtectedRoute;
