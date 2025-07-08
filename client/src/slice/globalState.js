import { createSlice } from "@reduxjs/toolkit";

const initialState = {
  recentProjects: JSON.parse(localStorage.getItem("_rp")) || [],
};

export const globalStateSlice = createSlice({
  name: "globalState",
  initialState,
  reducers: {
    // save the last project visited
    addRecentProjects: (state, action) => {
      state.recentProjects = JSON.parse(localStorage.getItem("_rp")) || [];
      if (state.recentProjects.includes(action.payload)) {
        state.recentProjects = state.recentProjects.filter((project) => project !== action.payload);
      }
      if (state.recentProjects.length >= 5) {
        state.recentProjects.pop();
      }
      state.recentProjects.unshift(action.payload);
      localStorage.setItem("_rp", JSON.stringify(state.recentProjects));
    },
    // clear the recent projects
    clearRecentProjects: (state) => {
      localStorage.removeItem("_rp");
      state.recentProjects = [];
    },
    // remove a project from the recent projects
    removeFromRecentProjects: (state, action) => {
      state.recentProjects = JSON.parse(localStorage.getItem("_rp")) || [];

      state.recentProjects = state.recentProjects.filter((project) => project !== action.payload);

      localStorage.setItem("_rp", JSON.stringify(state.recentProjects));
    },
  },
});

export const { addRecentProjects, clearRecentProjects, removeFromRecentProjects } = globalStateSlice.actions;

export default globalStateSlice.reducer;
