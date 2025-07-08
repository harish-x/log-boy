import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

import { Edit3, Save, X, KeyRound, Copy, Info, CheckCircle, AlertTriangle, CalendarDays, Power, FileText, Tag } from "lucide-react";
import { useGetProjectByNameQuery, useUpdateProjectMutation, useLazyGenerateProjectkeyQuery } from "@/services/ProjectService";

const Settings = () => {
  const { projectName } = useParams();
  const { data: ProjectResponse, isLoading: isLoadingProject, refetch } = useGetProjectByNameQuery(projectName);

  const [projectData, setProjectData] = useState();
  // Set project data
  useEffect(() => {
    setProjectData(ProjectResponse?.data);
  }, [ProjectResponse]);
  // Update project mutation
  const [updateProject, { isLoading: isUpdating }] = useUpdateProjectMutation();
  // Generate project key query
  const [generateToken, { isLoading: isGeneratingToken }] = useLazyGenerateProjectkeyQuery(projectData?.name);

  const [isEditing, setIsEditing] = useState(false); // State to track edit mode
  const [formData, setFormData] = useState({});
  const [generatedToken, setGeneratedToken] = useState(null);
  const [showTokenCopied, setShowTokenCopied] = useState(false);

  // Set form data
  useEffect(() => {
    if (projectData) {
      setFormData({
        description: projectData.description,
        active: projectData.active,
        retention_period: projectData.retention_period,
      });
    }
  }, [projectData]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  // Handle switch change for active and inactive
  const handleSwitchChange = (checked) => {
    setFormData((prev) => ({ ...prev, active: checked }));
  };

  // Handle edit toggle
  const handleEditToggle = () => {
    if (isEditing && projectData) {
      setFormData({
        name: projectData.name,
        description: projectData.description,
        active: projectData.active,
        retention_period: projectData.retention_period,
      });
    }
    setIsEditing(!isEditing);
  };

  // Handle save changes
  const handleSaveChanges = async () => {
    if (!projectData) return;
    try {
      await updateProject({ name: projectData.name, ...formData });
      toast.success("Project Updated", {
        description: "Your project settings have been saved.",
        action: <CheckCircle className="text-green-500" />,
      });
      setIsEditing(false);
      refetch();
    } catch (error) {
      toast.error("Update Failed", {
        description: "Could not save project settings. Please try again.",
        variant: "destructive",
        action: <AlertTriangle className="text-red-500" />,
      });
    }
  };
  // Handle token generation
  const handleGenerateToken = async () => {
    if (!projectData?.id) return;
    try {
      const result = await generateToken().unwrap();
      setGeneratedToken(result.data.key);
    } catch (error) {
      setGeneratedToken(null);
      toast.error("Token Generation Failed", {
        description: "Could not generate a new token. Please try again.",
        variant: "destructive",
        action: <AlertTriangle className="text-red-500" />,
      });
    }
  };

  // Handle token copy to clipboard
  const copyToClipboard = (text) => {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        setShowTokenCopied(true);
        setTimeout(() => setShowTokenCopied(false), 2000);
        toast.info("Token Copied!", {
          description: "The API token has been copied to your clipboard.",
        });
      })
      .catch((err) => {
        toast.warning("Copy Failed", {
          description: "Could not copy the token. Please try again manually.",
          variant: "destructive",
        });
      });
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  // Loading state
  if (isLoadingProject) {
    return (
      <div className="projects-container bg-background px-4 py-6 md:px-8 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] h-[calc(100vh-5rem)] overflow-y-auto">
        <Card className="max-w-3xl mx-auto">
          <CardHeader>
            <Skeleton className="h-8 w-3/4 mb-2" />
            <Skeleton className="h-4 w-1/2" />
          </CardHeader>
          <CardContent className="space-y-6">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="space-y-2">
                <Skeleton className="h-4 w-1/4" />
                <Skeleton className="h-10 w-full" />
              </div>
            ))}
          </CardContent>
          <CardFooter className="flex justify-end space-x-2">
            <Skeleton className="h-10 w-24" />
          </CardFooter>
        </Card>
        <Separator className="my-8 max-w-3xl mx-auto" />
        <Card className="max-w-3xl mx-auto">
          <CardHeader>
            <Skeleton className="h-6 w-1/2 mb-2" />
          </CardHeader>
          <CardContent className="space-y-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-1/3" />
          </CardContent>
        </Card>
      </div>
    );
  }
  // Error state
  if (!projectData) {
    return (
      <div className="projects-container flex items-center justify-center bg-background px-4 py-6 md:px-8 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] h-[calc(100vh-5rem)]">
        <Card className="max-w-md mx-auto text-center">
          <CardHeader>
            <CardTitle className="text-2xl flex items-center justify-center">
              <AlertTriangle className="w-8 h-8 mr-2 text-destructive" /> Project Not Found
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">The project "{projectName}" could not be found or loaded.</p>
            <p className="text-muted-foreground mt-2">Please check the project name or try again later.</p>
          </CardContent>
          <CardFooter>
            <Button onClick={() => window.history.back()} variant="outline" className="w-full">
              Go Back
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="projects-container bg-background px-4 py-6 md:px-8 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] h-[calc(100vh-5rem)] overflow-y-auto">
      <Card className="max-w-3xl mx-auto shadow-lg">
        <CardHeader>
          <div className="flex justify-between items-start">
            <div>
              <CardTitle className="text-2xl md:text-3xl flex items-center">
                <Info className="w-7 h-7 mr-3 text-primary" /> Project Settings
              </CardTitle>
              <CardDescription>Manage your project details and API access.</CardDescription>
            </div>
            {!isEditing && (
              <Button variant="outline" size="sm" onClick={handleEditToggle} disabled={isUpdating}>
                <Edit3 className="mr-2 h-4 w-4" /> Edit
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-6 pt-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 items-center">
            <Label className="md:text-right md:pr-4 flex items-center text-muted-foreground">
              <Tag className="w-4 h-4 mr-2" /> Project Name
            </Label>
            <div className="md:col-span-2">
              <p className="text-sm text-muted-foreground py-2 px-3 rounded-md bg-muted/50 break-all min-h-[2.5rem] flex items-center">
                {projectData.name}
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 items-start">
            <Label htmlFor="description" className="md:text-right md:pr-4 flex items-center text-muted-foreground pt-2">
              <FileText className="w-4 h-4 mr-2" /> Description
            </Label>
            <div className="md:col-span-2">
              {isEditing ? (
                <Textarea
                  id="description"
                  name="description"
                  value={formData.description || ""}
                  onChange={handleInputChange}
                  rows={4}
                  className="text-base"
                  disabled={isUpdating}
                />
              ) : (
                <p className="text-base py-2 px-3 rounded-md bg-muted min-h-[2.5rem] whitespace-pre-wrap">
                  {projectData.description || <span className="italic text-muted-foreground/70">No description provided.</span>}
                </p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 items-center">
            <Label htmlFor="active" className="md:text-right md:pr-4 flex items-center text-muted-foreground">
              <Power className="w-4 h-4 mr-2" /> Active Status
            </Label>
            <div className="md:col-span-2 flex items-center">
              {isEditing ? (
                <Switch id="active" checked={formData.active || false} onCheckedChange={handleSwitchChange} disabled={isUpdating} />
              ) : (
                <span
                  className={`py-2 px-3 rounded-md text-sm font-medium ${
                    projectData.active
                      ? "bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-400"
                      : "bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-400"
                  }`}
                >
                  {projectData.active ? "Active" : "Inactive"}
                </span>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 items-center">
            <Label htmlFor="retention_period" className="md:text-right md:pr-4 flex items-center text-muted-foreground">
              <CalendarDays className="w-4 h-4 mr-2" /> Retention Period
            </Label>
            <div className="md:col-span-2">
              {isEditing ? (
                <Input
                  id="retention_period"
                  name="retention_period"
                  value={formData.retention_period || ""}
                  onChange={handleInputChange}
                  className="text-base"
                  disabled={isUpdating}
                  placeholder="e.g., 30 DAYS, 2 MONTHS"
                />
              ) : (
                <p className="text-base font-medium py-2 px-3 rounded-md bg-muted min-h-[2.5rem] flex items-center">{projectData.retention_period}</p>
              )}
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 items-center">
            <Label className="md:text-right md:pr-4 flex items-center text-muted-foreground">
              <KeyRound className="w-4 h-4 mr-2" /> Project ID
            </Label>
            <div className="md:col-span-2">
              <p className="text-sm text-muted-foreground py-2 px-3 rounded-md bg-muted/50 break-all min-h-[2.5rem] flex items-center">
                {projectData.id}
              </p>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 items-center">
            <Label className="md:text-right md:pr-4 flex items-center text-muted-foreground">
              <CalendarDays className="w-4 h-4 mr-2" /> Created At
            </Label>
            <div className="md:col-span-2">
              <p className="text-sm text-muted-foreground py-2 px-3 rounded-md bg-muted/50 min-h-[2.5rem] flex items-center">
                {formatDate(projectData.created_at)}
              </p>
            </div>
          </div>
        </CardContent>
        {isEditing && (
          <CardFooter className="flex justify-end space-x-3 border-t pt-6">
            <Button variant="outline" onClick={handleEditToggle} disabled={isUpdating}>
              <X className="mr-2 h-4 w-4" /> Cancel
            </Button>
            <Button onClick={handleSaveChanges} disabled={isUpdating}>
              <Save className="mr-2 h-4 w-4" /> {isUpdating ? "Saving..." : "Save Changes"}
            </Button>
          </CardFooter>
        )}
      </Card>

      <Separator className="my-8 max-w-3xl mx-auto" />

      <Card className="max-w-3xl mx-auto shadow-lg">
        <CardHeader>
          <CardTitle className="text-2xl flex items-center">
            <KeyRound className="w-6 h-6 mr-3 text-primary" /> API Token
          </CardTitle>
          <CardDescription>Generate and manage your API access token for this project.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button onClick={handleGenerateToken} disabled={isGeneratingToken} className="w-full md:w-auto">
            <KeyRound className="mr-2 h-4 w-4" /> {isGeneratingToken ? "Generating Token..." : "Generate New Token"}
          </Button>
          {generatedToken && (
            <div className="space-y-2 pt-4">
              <Label htmlFor="apiToken" className="font-semibold">
                Your New API Token:
              </Label>
              <div className="flex items-center space-x-2">
                <Input id="apiToken" type="text" value={generatedToken} readOnly className="text-sm bg-muted flex-grow font-mono" />
                <Button variant="outline" size="icon" onClick={() => copyToClipboard(generatedToken)} title="Copy token">
                  {showTokenCopied ? <CheckCircle className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
              <p className="text-xs text-amber-600 dark:text-amber-500 flex items-center">
                <AlertTriangle className="w-3 h-3 mr-1.5 flex-shrink-0" />
                Please save this token securely. You will not be able to see it again.
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default Settings;
