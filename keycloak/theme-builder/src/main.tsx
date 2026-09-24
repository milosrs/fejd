import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { KcPage, type KcContext } from "./kc.gen";
import "./styles.css";

const kcContext = window.kcContext;

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        {kcContext === undefined ? (
            <h1>No Keycloak context provided</h1>
        ) : (
            <KcPage kcContext={kcContext as KcContext} />
        )}
    </StrictMode>
);
