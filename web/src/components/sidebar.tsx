import * as React from "react"
import { BookIcon, Check, ChevronsUpDown, PickaxeIcon, GithubIcon, InspectIcon, FileIcon, RailSymbol, Building2Icon, LucideBuilding, CogIcon, HistoryIcon, BarChartIcon, DoorOpenIcon, ScanEyeIcon, ListIcon } from "lucide-react"
import { Link, useLocation } from 'react-router-dom'

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Separator } from "@/components/ui/separator"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "./ui/card"
import { useEffect, useState } from "react"
import { useProject } from '@/contexts/project_context';
import { Project, useGetProjects } from '@/types'; // Assuming you have this hook from Orval
import { useConfig } from "@/contexts/config_context"
import { randomUUID } from "crypto"

interface SidebarProps {
  children: React.ReactNode;
}

export default function SidebarComponent({ children }: SidebarProps) {

  const { API_BASE_URL, WEBSOCKET_BASE_URL } = useConfig();

  const location = useLocation();
  const [currentPath, setCurrentPath] = useState<string[]>([])
  const { selectedProject, setSelectedProject } = useProject();
  const { data } = useGetProjects();
  const [projects, setProjects] = useState<Project[]>([]);

  useEffect(() => {
    if (data) {
      setProjects(data.data);
    }
  }, [data]);

  useEffect(() => {
    if (location.pathname === '/') {
      setCurrentPath([]);
      return;
    }

    const splitPath = location.pathname.split('/');
    // Remove the first element
    splitPath.shift();

    setCurrentPath(splitPath);
  }, [location.pathname]);

  const navData = {
    navMain: [
      {
        title: "Getting started",
        url: "#",
        items: [
          {
            title: "Projects",
            url: "/projects",
            isActive: true,
            disabled: false,
            icon: <Building2Icon />
          },
          {
            title: "Tasks",
            url: `/projects/${selectedProject}`,
            isActive: false,
            disabled: false,
            icon: <ListIcon />
          },
          {
            title: "Supervisors",
            url: "/supervisors",
            isActive: false,
            disabled: false,
            icon: <ScanEyeIcon />
          },
          {
            title: "Tools",
            url: `/tools`,
            isActive: false,
            disabled: false,
            icon: <PickaxeIcon />
          },
          {
            title: "Execution History",
            url: "/execution_history",
            isActive: false,
            disabled: true,
            icon: <HistoryIcon />
          },
          {
            title: "Export Data",
            url: "/export",
            isActive: false,
            disabled: true,
            icon: <DoorOpenIcon />
          },
          {
            title: "Stats",
            url: "/stats",
            isActive: false,
            disabled: true,
            icon: <BarChartIcon />
          },
          {
            title: "API Spec",
            url: "/api",
            isActive: false,
            disabled: false,
            icon: <CogIcon />
          },
          {
            title: "GitHub",
            url: "https://github.com/asteroidai/asteroid",
            isActive: false,
            disabled: false,
            icon: <GithubIcon />
          },
          {
            title: "Documentation",
            url: "https://docs.asteroid.ai",
            isActive: false,
            disabled: false,
            icon: <BookIcon />
          },
        ],
      }
    ],
  }

  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <SidebarMenuButton
                    size="lg"
                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                  >
                    <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                      <img src="/logo-128.png" alt="Asteroid" className="size-8 rounded-lg" />
                    </div>
                    <div className="flex flex-col gap-0.5 leading-none">
                      <span className="font-semibold">Asteroid Platform</span>
                      <span className="text-xs">
                        {projects && selectedProject ?
                          projects?.find(p => p.id === selectedProject)?.name || 'Select Project'
                          : 'Select Project'}
                      </span>
                    </div>
                    <ChevronsUpDown className="ml-auto" />
                  </SidebarMenuButton>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  className="w-[--radix-dropdown-menu-trigger-width]"
                  align="start"
                >
                  {projects && projects.map((project) => (
                    <DropdownMenuItem
                      key={project.id}
                      onSelect={() => setSelectedProject(project.id)}
                    >
                      {project.name}
                      {project.id === selectedProject && (
                        <Check className="ml-auto" />
                      )}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </SidebarMenuItem>
          </SidebarMenu>
          <form>
            <SidebarGroup className="py-0">
              <SidebarGroupContent className="relative">
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="text-xs tracking-wide">Agent Supervision & Evaluation</span>
                </div>
              </SidebarGroupContent>
            </SidebarGroup>
          </form>
        </SidebarHeader>
        <SidebarContent>
          {navData.navMain.map((item) => (
            <SidebarGroup key={item.title}>
              <SidebarGroupLabel>{item.title}</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {item.items.map((subItem) => (
                    <SidebarMenuItem key={subItem.title}>
                      <SidebarMenuButton
                        asChild
                        isActive={currentPath[currentPath.length - 1] === subItem.url}
                        disabled={subItem.disabled}
                        className={subItem.disabled ? "opacity-50 cursor-not-allowed" : ""}
                      >
                        <Link to={subItem.url} onClick={e => subItem.disabled && e.preventDefault()}>
                          {subItem.icon}
                          {subItem.title}
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          ))}
        </SidebarContent>

        <SidebarFooter>
          <div className="p-1">
            <Card className="shadow-none">
              <form>
                <CardHeader className="p-4 pb-0">
                  <CardTitle className="text-sm">
                    Config
                  </CardTitle>
                  <CardDescription>
                  </CardDescription>
                </CardHeader>
                <CardContent className="grid gap-2.5 p-4">
                  <p className="text-xs font-mono">[API] {API_BASE_URL}</p>
                  <div className="flex items-center gap-1">
                    <p className="text-xs font-mono">[WS] {WEBSOCKET_BASE_URL}</p>
                  </div>
                </CardContent>
              </form>
            </Card>
          </div>
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink asChild={true}>
                  <Link to="/">
                    home
                  </Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
              {currentPath.length > 0 && (
                currentPath.map((path, index) => (
                  <BreadcrumbItem key={index}>
                    <span className="text-muted-foreground text-xs">/</span>
                    <BreadcrumbLink asChild={true}>
                      <Link to={`/${currentPath.slice(0, index + 1).join('/')}`}>
                        {path}
                      </Link>
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                ))
              )}
            </BreadcrumbList>
          </Breadcrumb>
          <div className="ml-auto">
            <Card className="bg-muted px-3 py-1 shadow-none">
              <div className="flex items-center gap-2">
                <img src="/logo-128.png" alt="Asteroid" className="h-4 w-4 rounded-lg" />
                <span className="text-sm text-muted-foreground">
                  Selected Project: {` `}
                  {projects && selectedProject ?
                    projects?.find(p => p.id === selectedProject)?.name || 'No Project Selected'
                    : 'No Project Selected'}
                </span>
              </div>
            </Card>
          </div>
        </header>
        <div className="flex-grow">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
