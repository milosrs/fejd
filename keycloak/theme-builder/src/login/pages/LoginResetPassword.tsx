import type { PageProps } from "keycloakify/login/pages/PageProps";
import { kcSanitize } from "keycloakify/lib/kcSanitize";
import type { KcContext } from "../KcContext";
import type { I18n } from "../i18n";

export default function LoginResetPassword(props: PageProps<Extract<KcContext, { pageId: "login-reset-password.ftl" }>, I18n>) {
    const { kcContext, i18n, Template, doUseDefaultCss, classes } = props;

    const { msg, msgStr } = i18n;
    const { url, realm, auth, messagesPerField } = kcContext;

    return (
        <Template
            {...{ kcContext, i18n, doUseDefaultCss, classes }}
            displayInfo={true}
            displayMessage={!messagesPerField.existsError("username")}
            headerNode={msg("forgotPasswordTitle")}
            infoNode={
                <p className="auth-switch">
                    {realm.duplicateEmailsAllowed ? msgStr("emailInstructionUsername") : msgStr("emailInstruction")}
                </p>
            }
        >
            <form className="auth-form" action={url.loginAction} method="post">
                <div className="field">
                    <label htmlFor="username" className="field-label">
                        {msgStr("usernameOrEmail")}
                    </label>
                    <input
                        id="username"
                        name="username"
                        type="text"
                        className="field-input"
                        defaultValue={auth.attemptedUsername ?? ""}
                        autoFocus
                        autoComplete="username"
                        aria-invalid={messagesPerField.existsError("username")}
                    />
                    {messagesPerField.existsError("username") && (
                        <p className="field-error">
                            <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.get("username")) }} />
                        </p>
                    )}
                </div>

                <button type="submit" className="btn btn-primary btn-block">
                    {msgStr("doSubmit")}
                </button>

                <p className="auth-switch">
                    <a className="auth-link" href={url.loginUrl}>
                        {msgStr("backToLogin")}
                    </a>
                </p>
            </form>
        </Template>
    );
}
