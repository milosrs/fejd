import { useState } from "react";
import type { PageProps } from "keycloakify/login/pages/PageProps";
import { kcSanitize } from "keycloakify/lib/kcSanitize";
import type { KcContext } from "../KcContext";
import type { I18n } from "../i18n";

const ROLES = ["Customer", "Employee", "Owner"] as const;

export default function Register(props: PageProps<Extract<KcContext, { pageId: "register.ftl" }>, I18n>) {
    const { kcContext, i18n, Template, doUseDefaultCss, classes } = props;

    const { msg, msgStr } = i18n;
    const { url, messagesPerField } = kcContext;

    // Invite flow: the role is carried into the registration URL and locks the
    // dropdown. A disabled <select> is not submitted, so the value is echoed
    // through a hidden input.
    const [inviteRole] = useState<string | undefined>(() => {
        const role = new URLSearchParams(window.location.search).get("registration_role");
        return role !== null && (ROLES as readonly string[]).includes(role) ? role : undefined;
    });

    return (
        <Template
            {...{ kcContext, i18n, doUseDefaultCss, classes }}
            displayMessage={
                !messagesPerField.existsError("firstName", "lastName", "email", "username", "password", "password-confirm")
            }
            headerNode={msg("registerTitle")}
            displayInfo={true}
            infoNode={
                <p className="auth-switch">
                    {msgStr("alreadyHaveAccount")}{" "}
                    <a className="auth-link" href={kcContext.url.loginUrl}>
                        {msgStr("doLogIn")}
                    </a>
                </p>
            }
        >
            <form className="auth-form" action={url.registrationAction} method="post">
                <div className="field">
                    <label htmlFor="firstName" className="field-label">
                        {msgStr("firstName")}
                    </label>
                    <input id="firstName" name="firstName" type="text" className="field-input" autoComplete="given-name" />
                    {messagesPerField.existsError("firstName") && (
                        <p className="field-error">
                            <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.get("firstName")) }} />
                        </p>
                    )}
                </div>

                <div className="field">
                    <label htmlFor="lastName" className="field-label">
                        {msgStr("lastName")}
                    </label>
                    <input id="lastName" name="lastName" type="text" className="field-input" autoComplete="family-name" />
                    {messagesPerField.existsError("lastName") && (
                        <p className="field-error">
                            <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.get("lastName")) }} />
                        </p>
                    )}
                </div>

                <div className="field">
                    <label htmlFor="email" className="field-label">
                        {msgStr("email")}
                    </label>
                    <input id="email" name="email" type="email" className="field-input" autoComplete="email" />
                    {messagesPerField.existsError("email") && (
                        <p className="field-error">
                            <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.get("email")) }} />
                        </p>
                    )}
                </div>

                <div className="field">
                    <label htmlFor="username" className="field-label">
                        {msgStr("username")}
                    </label>
                    <input id="username" name="username" type="text" className="field-input" autoComplete="username" />
                    {messagesPerField.existsError("username") && (
                        <p className="field-error">
                            <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.get("username")) }} />
                        </p>
                    )}
                </div>

                <div className="field">
                    <label htmlFor="role" className="field-label">
                        {msgStr("registerRoleLabel")}
                    </label>
                    {inviteRole !== undefined ? (
                        <>
                            <select id="role" className="field-input field-select" disabled value={inviteRole}>
                                <option value="Customer">{msgStr("roleCustomer")}</option>
                                <option value="Employee">{msgStr("roleEmployee")}</option>
                                <option value="Owner">{msgStr("roleOwner")}</option>
                            </select>
                            <input type="hidden" name="user.attributes.registration_role" value={inviteRole} />
                        </>
                    ) : (
                        <select
                            id="role"
                            name="user.attributes.registration_role"
                            className="field-input field-select"
                            defaultValue="Customer"
                        >
                            <option value="Customer">{msgStr("roleCustomer")}</option>
                            <option value="Employee">{msgStr("roleEmployee")}</option>
                            <option value="Owner">{msgStr("roleOwner")}</option>
                        </select>
                    )}
                </div>

                {kcContext.passwordRequired && (
                    <>
                        <div className="field">
                            <label htmlFor="password" className="field-label">
                                {msgStr("password")}
                            </label>
                            <input id="password" name="password" type="password" className="field-input" autoComplete="new-password" />
                            {messagesPerField.existsError("password") && (
                                <p className="field-error">
                                    <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.get("password")) }} />
                                </p>
                            )}
                        </div>

                        <div className="field">
                            <label htmlFor="password-confirm" className="field-label">
                                {msgStr("passwordConfirm")}
                            </label>
                            <input id="password-confirm" name="password-confirm" type="password" className="field-input" autoComplete="new-password" />
                            {messagesPerField.existsError("password-confirm") && (
                                <p className="field-error">
                                    <span dangerouslySetInnerHTML={{ __html: kcSanitize(messagesPerField.get("password-confirm")) }} />
                                </p>
                            )}
                        </div>
                    </>
                )}

                <button type="submit" className="btn btn-primary btn-block">
                    {msgStr("doRegister")}
                </button>
            </form>
        </Template>
    );
}
