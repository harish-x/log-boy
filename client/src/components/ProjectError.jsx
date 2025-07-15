import React from "react";

const ProjectError = (projectName) => {
  return (
    <div className="flex items-center justify-center h-[calc(100vh-5rem)] mx-auto rounded-2xl border border-primary/[0.20]">
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
};

export default ProjectError;
