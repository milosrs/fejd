import type { PageProps } from "keycloakify/login/pages/PageProps";
import { kcSanitize } from "keycloakify/lib/kcSanitize";
import type { KcContext } from "../KcContext";
import type { I18n } from "../i18n";

export default function Login(props: PageProps<Extract<KcContext, { pageId: "login.ftl" }>, I18n>) {
    const { kcContext, i18n, Template, doUseDefaultCss, classes } = props;

    const { msg, msgStr } = i18n;
    const { realm, url, login, messagesPerField } = kcContext;

    return (
        <Template
            {...{ kcContext, i18n, doUseDefaultCss, classes }}
            displayMessage={!messagesPerField.existsError("username", "password")}
            headerNode={msg("loginTitle")}
            displayInfo={realm.password && realm.registrationAllowed && !kcContext.registrationDisabled}
            infoNode={
                <p className="auth-switch">
                    {msgStr("noAccount")}{" "}
                    <a className="auth-link" href={url.registrationUrl}>
                        {msgStr("doRegister")}
                    </a>
                </p>
            }
        >
            <form className="auth-form" action={url.loginAction} method="post">
                <p className="auth-subtitle">{msgStr("loginIntro")}</p>

                <div className="field">
                    <label htmlFor="username" className="field-label">
                        {msgStr("usernameOrEmail")}
                    </label>
                    <input
                        id="username"
                        name="username"
                        type="text"
                        className="field-input"
                        defaultValue={login.username ?? ""}
                        autoFocus
                        autoComplete="username"
                        aria-invalid={messagesPerField.existsError("username", "password")}
                    />
                    {messagesPerField.existsError("username", "password") && (
                        <p className="field-error">
                            <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.getFirstError("username", "password")) }} />
                        </p>
                    )}
                </div>

                <div className="field">
                    <label htmlFor="password" className="field-label">
                        {msgStr("password")}
                    </label>
                    <input id="password" name="password" type="password" className="field-input" autoComplete="current-password" />
                </div>

                <div className="auth-options">
                    {realm.resetPasswordAllowed && (
                        <a className="auth-link" href={url.loginResetCredentialsUrl}>
                            {msgStr("forgotPassword")}
                        </a>
                    )}
                </div>

                <button type="submit" className="btn btn-primary btn-block">
                    {msgStr("doLogIn")}
                </button>
            </form>
        </Template>
    );
}
