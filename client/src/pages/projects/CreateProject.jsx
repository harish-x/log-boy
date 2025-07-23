import { Separator } from "@radix-ui/react-dropdown-menu";
import { Eye, Clipboard, ClipboardCheck, ArrowLeft, OctagonAlert } from "lucide-react";
import React from "react";
import { useCreateProjectMutation } from "../../services/ProjectService";
import { toast } from "sonner";
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select";

const CreateProject = () => {
  const [projectName, setProjectName] = React.useState("");
  const [projectDescription, setProjectDescription] = React.useState("");
  const [logsRetentionPeriod, setLogsRetentionPeriod] = React.useState("");
  const [created, setCreated] = React.useState(true);

  const [nameError, setNameError] = React.useState("");

  const [copiedProjectName, setCopiedProjectName] = React.useState(false);
  const [copiedKey, setCopiedKey] = React.useState(false);

  const [createProject, { isLoading, error, data: projectResponse }] = useCreateProjectMutation();

  // Validate project name
  const validateProjectName = (name) => {
    if (!name) {
      setNameError("Project name is required");
      return false;
    }

    if (name.length < 3 || name.length > 30) {
      setNameError("Project name must be between 3 and 30 characters");
      return false;
    }

    if (!/^[a-z]/.test(name)) {
      setNameError("Project name must start with a lowercase letter");
      return false;
    }

    if (!/^[a-z][a-z0-9_]*$/.test(name)) {
      setNameError("Project name can only contain lowercase letters, numbers, and underscores");
      return false;
    }

    setNameError("");
    return true;
  };

  // Handle project name change
  const handleProjectNameChange = (e) => {
    const value = e.target.value;
    setProjectName(value);
    validateProjectName(value);
  };

  // Handle form submission
  const handleSubmit = (e) => {
    e.preventDefault();
    if (validateProjectName(projectName)) {
      createProject({
        name: projectName,
        description: projectDescription,
        retention_period: logsRetentionPeriod,
      })
        .unwrap()
        .then(() => {
          toast.success("Project created successfully!", { duration: 5000 });
          setCreated(true); // set created to true to render the project page
        })
        .catch((error) => {
          toast.error(`${error?.data?.message}`, { duration: 5000 });
          setNameError(error?.data?.message || "An error occurred");
        })
        .finally(() => {
          // reset the form
          setProjectName("");
          setProjectDescription("");
          setLogsRetentionPeriod("");
        });
    }
  };

  return (
    <div className="w-[98%] mx-auto rounded-2xl border border-primary/[0.10] h-[calc(100vh-5rem)] bg-background">
      {!created ? ( // before project created
        <div className="flex flex-col items-center justify-center h-full">
          __
          <h1 className="text-2xl font-bold text-primary mb-4">Create New Project</h1>
          <form className="w-full max-w-md">
            <div className="mb-4">
              <label className="block text-sm font-medium text-primary mb-2" htmlFor="project-name">
                Project Name
              </label>
              <input
                type="text"
                id="project-name"
                className={`w-full px-3 py-2 border b rounded-lg focus:outline-none focus:ring-0 ${
                  nameError ? "border-destructive focus:destructive" : "border-primary/[0.20] focus:ring-primary"
                }`}
                placeholder="Enter project name"
                onChange={handleProjectNameChange}
                required
              />
              {nameError && <p className="mt-1 text-sm text-destructive">{nameError}</p>}
              {projectName && !nameError && <p className="mt-1 text-sm text-green-600">✓ Valid project name</p>}
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-primary mb-2" htmlFor="project-description">
                Description
              </label>
              <textarea
                id="project-description"
                className="w-full px-3 py-2 border border-primary/[0.20] rounded-lg focus:outline-none"
                placeholder="Enter project description"
                rows="4"
                maxLength={200}
                style={{ resize: "none" }}
                onChange={(e) => setProjectDescription(e.target.value)}
              ></textarea>
            </div>
            <div className="mb-4">
              <Select value={logsRetentionPeriod} onValueChange={setLogsRetentionPeriod}>
                <SelectTrigger className="w-full px-3 py-2 border border-primary/[0.20] rounded-lg focus:outline-none focus:ring-2 focus:ring-primary">
                  <SelectValue placeholder="Select logs retention period" />
                </SelectTrigger>
                <SelectContent>
                  {Array.from({ length: 12 }, (_, i) => (
                    <SelectItem key={i} value={`${i + 1} MONTHS`}>
                      {i + 1} month{i + 1 > 1 ? "s" : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <button
              type="submit"
              onClick={handleSubmit}
              className="w-full bg-primary text-accent py-2 rounded-lg hover:bg-primary/[0.80] transition duration-200 cursor-pointer font-medium"
              disabled={isLoading}
            >
              {isLoading ? "Creating..." : "Create Project"}
            </button>
          </form>
        </div>
      ) : (
        // after project created
        <div className="flex flex-col items-center justify-center h-full">
          <div className="flex items-center justify-between w-full max-w-xl mb-6">
            <button
              className=" mb-4 text-primary hover:text-primary/[0.80] transition duration-200 cursor-pointer rounded-full bg-accent/[0.10] p-2 flex items-center"
              type="button"
              aria-label="Back to create project"
              onClick={() => setCreated(false)}
            >
              <ArrowLeft size={24} />
            </button>
          </div>

          <h2 className="text-xl font-bold text-primary mb-4">Project Details</h2>
          <p className="text-sm text-muted-foreground mb-2">
            Your project has been created successfully! You can now use the following details to access your project.
          </p>

          <div className="mb-4 w-full max-w-xl flex items-center mt-6">
            <input
              type="text"
              readOnly
              value={!isLoading && projectResponse?.data?.name}
              id="project-name"
              className="flex-1 px-3 py-2 border border-primary/[0.20] rounded-l-lg bg-background text-primary focus:outline-none"
            />
            <button
              type="button"
              className="px-3 py-3 bg-primary dark:bg-accent text-accent rounded-r-lg transition duration-200 cursor-pointer flex items-center"
              onClick={() => {
                navigator.clipboard.writeText(!isLoading && projectResponse?.data?.name); // copy to clipboard
                setCopiedProjectName(true);
                setTimeout(() => setCopiedProjectName(false), 1500);
              }}
            >
              {copiedProjectName ? <ClipboardCheck size={19} color="#22c55e" /> : <Clipboard color="#fff" size={19} />}
            </button>
          </div>
          <div className="mb-4 w-full max-w-xl flex items-center">
            <input
              type="password"
              readOnly
              value={!isLoading && projectResponse?.data?.key}
              id="project-key"
              className="flex-1 px-3 py-2 border border-primary/[0.20] rounded-l-lg bg-background text-primary focus:outline-none"
            />
            <button
              type="button"
              className="px-3 py-3 bg-primary dark:bg-accent text-accent  transition duration-200 cursor-pointer"
              onClick={() => {
                const input = document.getElementById("project-key");
                input.type = input.type === "password" ? "text" : "password";
                input.focus();
              }}
            >
              <Eye color="#fff" size={19} />
            </button>
            <Separator orientation="vertical" className="border border-primary/[0.10]" />
            <button
              type="button"
              className="px-3 py-3 bg-primary dark:bg-accent text-accent rounded-r-lg transition duration-200 cursor-pointer flex items-center"
              onClick={() => {
                navigator.clipboard.writeText(!isLoading && projectResponse?.data?.key);
                setCopiedKey(true);
                setTimeout(() => setCopiedKey(false), 1500);
              }}
            >
              {copiedKey ? <ClipboardCheck size={19} color="#22c55e" /> : <Clipboard size={19} color="#fff" />}
            </button>
          </div>
          <span className="flex items-center text-sm opacity-70">
            <OctagonAlert className="mr-2" /> This is the only time you will be able to see this information
          </span>
        </div>
      )}
    </div>
  );
};

export default CreateProject;
