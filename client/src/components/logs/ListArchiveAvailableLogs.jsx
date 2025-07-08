import { useEffect, useRef, useState } from "react";
import { Skeleton } from "../ui/skeleton";
import { ContextMenu, ContextMenuContent, ContextMenuItem, ContextMenuTrigger } from "@/components/ui/context-menu";
import { useNavigate } from "react-router-dom";
import { useParams } from "react-router-dom";

const ListArchiveAvailableLogs = ({ isLoading, isError, isFetching, data: availableLogs }) => {
  const [selected, setSelected] = useState(null);

  const logContainerRef = useRef(null);
  const navigate = useNavigate();
  const { projectName } = useParams();
  useEffect(() => {
    function handleClickOutside(event) {
      if (logContainerRef.current && !logContainerRef.current.contains(event.target)) {
        setSelected(null);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);
  if (isLoading || isFetching) {
    return (
      <div className="flex gap-2 p-10 flex-wrap">
        {Array.from({ length: 10 }).map((_, index) => (
          <Skeleton key={index} className="min-h-36 w-36 " />
        ))}
      </div>
    );
  }
  if (availableLogs?.data?.logs?.length === 0 || !availableLogs?.data?.logs) {
    return (
      <div className="p-10">
        <p className="text-center text-2xl">No logs available yet 👀</p>
      </div>
    );
  }
  function handlenavigate(fileName) {
    navigate(`archive/${encodeURIComponent(fileName)}`);
    setSelected(null);
  }
  return (
    <div className="p-10">
      <div className="flex gap-3" ref={logContainerRef}>
        {availableLogs.data.logs.map((log, index) => (
          <ContextMenu>
            <ContextMenuTrigger>
              <div
                key={index}
                onClick={() => setSelected(index)}
                onDoubleClick={() => handlenavigate(log)}
                className={`cursor-pointer p-2 rounded text-center flex flex-col items-center justify-center ${
                  selected === index ? "bg-primary/[0.20] dark:bg-accent" : ""
                } ${selected !== index ? "hover:bg-primary/[0.10] dark:hover:bg-accent/50" : ""}`}
              >
                <img src="/log.png" alt="" className="w-28" />
                <p className="text-center text-sm">{log}</p>
              </div>
            </ContextMenuTrigger>
            <ContextMenuContent>
              <ContextMenuItem onClick={() => handlenavigate(log)}>Open</ContextMenuItem>
            </ContextMenuContent>
          </ContextMenu>
        ))}
      </div>
    </div>
  );
};

export default ListArchiveAvailableLogs;
