import React, { useEffect } from "react";
import Login from "@/components/Login";
import { useNavigate } from "react-router-dom";
import { useMsal } from "@azure/msal-react";
const Home = () => {
  const { accounts, inProgress } = useMsal();
  const navigate = useNavigate();
  useEffect(() => {
    if (inProgress === "none" && accounts.length > 0) {
      navigate("/dashboard/projects", { replace: true });
    }
  }, [accounts, inProgress, navigate]);

  if (inProgress !== "none") {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <span className="loader"></span>
      </div>
    );
  }
  return (
    <div>
      <Login />
    </div>
  );
};

export default Home;
