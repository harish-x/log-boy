import { configureStore } from "@reduxjs/toolkit";
import { setupListeners } from "@reduxjs/toolkit/query";
import globalReducer from "@/slice/globalState";
import { reAuthQuery } from "./services/BaseQuery";

export const store = configureStore({
  reducer: {
    globalState: globalReducer,
    [reAuthQuery.reducerPath]: reAuthQuery.reducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(reAuthQuery.middleware),
});

setupListeners(store.dispatch);
