import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Dashboard from "./pages/Dashboard";
import Projects from "./pages/projects/Projects";
import { ThemeProvider } from "@/components/theme-provider";
import CreateProject from "./pages/projects/CreateProject";
import { Toaster } from "@/components/ui/sonner";
import Project from "./pages/projects/Project";
import Statistics from "./pages/logs/Statistics";
import LiveWatch from "./pages/logs/LiveWatch";
import ProjectSettings from "@/pages/projects/Settings";
import ProtectedRoute from "@/components/ProtectedRoute";
import PageNotFound from "./pages/PageNotFound";
import Home from "@/pages/home/Home";
import "./App.css";
import ColdStorage from "@/pages/logs/ColdStorage";
import ArchiveLogs from "@/components/logs/ArchiveLogs";
import Monitoring from "@/pages/metrics/Monitoring";
import AlertManager from "@/pages/alerts/Alert";
import Alerts from "@/pages/alerts/Alerts";

const App = () => {
  return (
    <>
      <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
        <Toaster />
        <Router>
          <Routes>
            <Route path="/" element={<Home />} />

            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <Dashboard />
                </ProtectedRoute>
              }
            >
              <Route path="projects" element={<Projects />} />
              <Route path="projects/create" element={<CreateProject />} />
              <Route path="project/:projectName" element={<Project />} />
              <Route path="project/:projectName/statistics" element={<Statistics />} />
              <Route path="project/:projectName/cold_storage" element={<ColdStorage />} />
              <Route path="project/:projectName/cold_storage/archive/:logName" element={<ArchiveLogs />} />
              <Route path="project/:projectName/live-watch" element={<LiveWatch />} />
              <Route path="project/:projectName/settings" element={<ProjectSettings />} />
              <Route path="project/:projectName/metrics" element={<Monitoring />} />
              <Route path="project/:projectName/alert/manage" element={<AlertManager />} />
              <Route path="project/:projectName/alert" element={<Alerts />} />
            </Route>

            <Route path="*" element={<PageNotFound />} />
          </Routes>
        </Router>
      </ThemeProvider>
    </>
  );
};

export default App;
