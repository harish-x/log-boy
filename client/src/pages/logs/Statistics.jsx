import Stats from "@/components/logs/Stats";
import React from "react";
import { useLazyGenerateStatsQuery, useGetProjectByNameQuery } from "@/services/ProjectService";
import { useParams } from "react-router-dom";
import { toast } from "sonner";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";

const Statistics = () => {
  const { projectName } = useParams();
  const {
    isLoading: isLoadingProject,
    isError: isErrorProject,
    isFetching: isFetchingProject,
    isSuccess: isSuccessProject,
  } = useGetProjectByNameQuery(projectName);
  const [generateStats] = useLazyGenerateStatsQuery();

  const [data, setData] = React.useState([]);

  // Get Stats data for the project
  React.useEffect(() => {
    function triggerGenerateStats() {
      generateStats(projectName)
        .unwrap()
        .then((response) => {
          setData(response.data);
        })
        .catch((error) => {
          toast.error(`${error?.data?.message}`);
        });
    }

    if (isSuccessProject && projectName) {
      triggerGenerateStats();
    }
  }, [projectName, isSuccessProject]);

  console.log(data);
  // if project not found
  if (isErrorProject) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-5rem)]">
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
    <div className="w-full p-2">
      <Stats profile={data} isFetchingProject={isFetchingProject} isLoadingProject={isLoadingProject} />
    </div>
  );
};

export default Statistics;
