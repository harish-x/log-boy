import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.jsx";
import { Provider } from "react-redux";
import { store } from "./store";
import { MsalProvider } from "@azure/msal-react";
import { msalInstance } from "./authConfig";

async function renderApp() {
  await msalInstance.initialize();

  createRoot(document.getElementById("root")).render(
    <MsalProvider instance={msalInstance}>
      <Provider store={store}>
        <App />
      </Provider>
    </MsalProvider>
  );
}

renderApp();
