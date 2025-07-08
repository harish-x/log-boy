import { reAuthQuery } from "./BaseQuery";

export const ProjectService = reAuthQuery.injectEndpoints({
  endpoints: (builder) => ({
    // get all projects
    getProjects: builder.query({
      query: (filter) => `projects?page=${filter.page}&limit=${filter.limit}`,
      invalidatesTags: ["Projects"],
      providesTags: (result) =>
        result
          ? [
              ...result.data.projects.map(({ id }) => ({
                type: "Projects",
                id,
              })),
              { type: "Projects", id: "LIST" },
            ]
          : [{ type: "Projects", id: "LIST" }],
    }),
    // create new project
    createProject: builder.mutation({
      query: (projectData) => ({
        url: "projects",
        method: "POST",
        body: projectData,
      }),
      invalidatesTags: ["Projects"],
    }),

    // get project by name
    getProjectByName: builder.query({
      query: (name) => `projects/${name}`,
      providesTags: (result, error, name) => [{ type: "Projects", id: name }],
    }),
    // delete project
    deleteProject: builder.mutation({
      query: (name) => ({
        url: `projects/${name}`,
        method: "DELETE",
      }),
      invalidatesTags: ["Projects"],
    }),
    // generate stats
    generateStats: builder.query({
      query: (name) => `projects/${name}/logs/stats`,
      providesTags: (result, error, name) => [{ type: "Projects", id: name }],
    }),
    // update project
    updateProject: builder.mutation({
      query: (projectData) => ({
        url: `projects/${projectData.name}`,
        method: "PUT",
        body: projectData,
      }),
      invalidatesTags: ["Projects"],
    }),
    // generate project key
    generateProjectkey: builder.query({
      query: (name) => `projects/${name}/key`,
    }),
    // get recent projects
    getRecentProjects: builder.query({
      query: (projectNames) => `projects/recent/projects?p=${projectNames.p}`,
    }),
  }),
});

export const {
  useGetProjectsQuery,
  useLazyGetProjectsQuery,
  useCreateProjectMutation,
  useDeleteProjectMutation,
  useGetProjectByNameQuery,
  useLazyGenerateStatsQuery,
  useUpdateProjectMutation,
  useLazyGenerateProjectkeyQuery,
  useLazyGetRecentProjectsQuery,
} = ProjectService;
