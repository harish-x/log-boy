import React from 'react'

const ProjectNotFound = (projectName) => {
  return (
    <div className="flex items-center justify-center h-[calc(100vh-5rem)]">
      <Card className="max-w-md mx-auto text-center">
        <CardHeader>
          <CardTitle className="text-2xl flex items-center justify-center">
            <AlertTriangle className="w-8 h-8 mr-2 text-destructive" /> Error
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">An error occurred while loading the project "{projectName}".</p>
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

export default ProjectNotFound
