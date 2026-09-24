import { useEffect } from "react";
import type { TemplateProps } from "keycloakify/login/TemplateProps";
import { kcSanitize } from "keycloakify/lib/kcSanitize";
import { PUBLIC_URL } from "keycloakify/PUBLIC_URL";
import type { KcContext } from "./KcContext";
import type { I18n } from "./i18n";

export default function Template(props: TemplateProps<KcContext, I18n>) {
    const {
        kcContext,
        i18n,
        children,
        displayInfo = false,
        displayMessage = true,
        headerNode,
        infoNode = null,
        socialProvidersNode = null,
        documentTitle,
        bodyClassName,
    } = props;

    const { msgStr, currentLanguage, enabledLanguages } = i18n;
    const { message, isAppInitiatedAction } = kcContext;

    const appUrl = kcContext.properties?.fejdAppUrl || "https://fejd.fyi";

    const langShort: Record<string, string> = { en: "EN", sr: "SR" };

    useEffect(() => {
        document.title = documentTitle ?? msgStr("loginTitle");
    }, [documentTitle, msgStr]);

    return (
        <div className={bodyClassName ?? "fejd-auth"}>
            <header className="topbar">
                <div className="topbar-inner">
                    <a className="topbar-brand" href={appUrl} aria-label={msgStr("backToFejd")}>
                        <img src={`${PUBLIC_URL}/logo_dark.jpg`} alt="fejd" className="brand-logo" />
                    </a>
                    <nav className="topbar-actions">
                        {enabledLanguages.length > 1 && (
                            <select
                                className="lang-select"
                                aria-label="Language"
                                value={currentLanguage.languageTag}
                                onChange={(event) => {
                                    const lang = enabledLanguages.find((l) => l.languageTag === event.target.value);
                                    if (lang) window.location.href = lang.href;
                                }}
                            >
                                {enabledLanguages.map((lang) => (
                                    <option key={lang.languageTag} value={lang.languageTag}>
                                        {langShort[lang.languageTag] ?? lang.label}
                                    </option>
                                ))}
                            </select>
                        )}
                        {kcContext.pageId === "login.ftl" ? (
                            <a className="btn btn-primary" href={kcContext.url.registrationUrl}>
                                {msgStr("doRegister")}
                            </a>
                        ) : (
                            <a className="btn btn-primary" href={kcContext.url.loginUrl}>
                                {msgStr("doLogIn")}
                            </a>
                        )}
                    </nav>
                </div>
            </header>

            <main className="auth-shell">
                <div className="auth-hero">
                    <img src={`${PUBLIC_URL}/logo_dark.jpg`} alt="fejd" className="hero-logo" />
                    <p className="hero-text">{msgStr("homeIntro")}</p>
                </div>

                <div className="card">
                    {displayMessage && message !== undefined && (message.type !== "warning" || !isAppInitiatedAction) && (
                        <div className={`alert alert-${message.type}`}>
                            <span className="alert-text" dangerouslySetInnerHTML={{ __html: kcSanitize(message.summary) }} />
                        </div>
                    )}

                    {headerNode !== null && headerNode !== undefined && (
                        <h1 className="auth-title">{headerNode}</h1>
                    )}

                    {children}

                    {socialProvidersNode !== null && socialProvidersNode !== undefined && socialProvidersNode}

                    {displayInfo && infoNode !== null && infoNode !== undefined && (
                        <div className="auth-info">{infoNode}</div>
                    )}
                </div>
            </main>
        </div>
    );
}
