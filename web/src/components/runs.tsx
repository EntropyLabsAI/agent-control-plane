import { useGetProjects, Project, useGetTaskRuns, Run, useGetProject, useGetRunTools, Tool } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import Page from "./util/page";
import { useProject } from "@/contexts/project_context";
import { UUIDDisplay } from "./util/uuid_display";
import { Button } from "./ui/button";
import { ArrowRightIcon, RailSymbol } from "lucide-react";
import { CreatedAgo } from "./util/created_ago";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableFooter,
} from "@/components/ui/table"
import { ProjectBadge, StatusBadge, SupervisionResultBadge, TaskBadge, ToolBadge, ToolBadges } from "./util/status_badge";
import SelectResult from "./select_result";

export default function Runs() {
  const [runs, setRuns] = useState<Run[]>([]);
  const { taskId } = useParams();
  const { selectedProject } = useProject();
  const [project, setProject] = useState<Project | undefined>(undefined);

  const { data: runsData, isLoading: runsLoading, error: runsError } = useGetTaskRuns(taskId || '');
  const { data: projectData, isLoading: projectLoading, error: projectError } = useGetProject(selectedProject || '');

  useEffect(() => {
    if (runsData?.data) {
      setRuns(runsData.data);
    } else {
      setRuns([]);
    }
  }, [runsData]);

  useEffect(() => {
    if (projectData?.data) {
      setProject(projectData.data);
    }
  }, [projectData]);

  return (
    <Page title={`Runs`}
      subtitle={<span>{runs.length > 0 ? `${runs.length} runs` : 'No runs'} found for <TaskBadge taskId={taskId ?? ''} /></span>}
      icon={<RailSymbol className="w-6 h-6" />}
    >
      {runs.length === 0 &&
        <p className="text-sm text-gray-500">No runs found for this project. When you run an agent, it will appear here.</p>
      }
      {runs.length > 0 && (
        // Display the tools of the first run 

        <div className="col-span-3 flex flex-col gap-4">
          <ToolsBadgeList runId={runs[0].id} />
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">Agent Run</TableHead>
                <TableHead className="w-[100px] text-right">Created</TableHead>
                <TableHead className="w-[100px] text-right">Status</TableHead>
                <TableHead className="w-[100px] text-right">Result</TableHead>
                <TableHead className="w-[50px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {runs.map((run) => (
                <TableRow key={run.id}>
                  <TableCell className="font-medium">
                    {runs.indexOf(run) + 1}
                  </TableCell>
                  <TableCell className="text-right">
                    <CreatedAgo datetime={run.created_at} label='' />
                  </TableCell>
                  <TableCell className="text-right">
                    <StatusBadge status={run.status} />
                  </TableCell>
                  <TableCell className="text-right">
                    {project && project.run_result_tags && run.status === 'completed' &&
                      <SelectResult
                        result={run.result}
                        possibleResults={project.run_result_tags}
                        runId={run.id}
                      />
                    }
                    {project && project.run_result_tags && run.status !== 'completed' &&
                      <p className="text-sm text-muted-foreground">
                        {run.status}
                      </p>
                    }
                  </TableCell>
                  <TableCell className="text-right">
                    <Link to={`/tasks/${taskId}/runs/${run.id}`}>
                      <Button variant="ghost"><ArrowRightIcon className="h-4 w-4" /></Button>
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TableCell className="text-xs text-muted-foreground" colSpan={7}>
                  {runs.length} runs found for this agent
                </TableCell>
              </TableRow>
            </TableFooter>
          </Table>
        </div>
      )
      }
    </Page >
  )
}

function ToolsBadgeList({ runId }: { runId: string }) {
  const { data: toolsData, isLoading: toolsLoading, error: toolsError } = useGetRunTools(runId || '');
  const [tools, setTools] = useState<Tool[]>([]);

  function deduplicateTools(tools: Tool[]) {
    // First, sort by name
    tools.sort((a, b) => a.name.localeCompare(b.name));

    return tools.filter((tool, index, self) =>
      index === self.findIndex((t) => t.id === tool.id)
    );
  }

  useEffect(() => {
    if (toolsData?.data) {
      setTools(deduplicateTools(toolsData.data));
    } else {
      setTools([]);
    }
  }, [toolsData]);

  return (
    <div className="flex flex-row gap-2 min-w-0">
      {toolsLoading && <p>Loading...</p>}
      {toolsError && <p>Error: {toolsError.message}</p>}
      <div className="flex-shrink-0">
        <ToolBadges tools={tools} maxTools={4} />
      </div>
    </div>
  )
}
