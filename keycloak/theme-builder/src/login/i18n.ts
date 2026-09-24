import { i18nBuilder } from "keycloakify/login";

const { useI18n, ofTypeI18n } = i18nBuilder
    .withExtraLanguages({
        // Serbian (Latin) is not in keycloakify's default language set. The base
        // default-set messages fall back to English (see sr.ts); every key the
        // theme actually renders is overridden below in both `en` and `sr`.
        sr: {
            label: "srpski",
            getMessages: () => import("./sr"),
        },
    })
    .withCustomTranslations({
        en: {
            loginTitle: "Sign in to Fejd",
            registerTitle: "Create your Fejd account",
            forgotPasswordTitle: "Reset your password",
            loginIntro: "Sign in to book appointments and manage your salon.",
            usernameOrEmail: "Email or username",
            forgotPassword: "Forgot password?",
            alreadyHaveAccount: "Already have an account?",
            noAccount: "Don't have an account yet?",
            homeIntro: "Book haircut appointments. Open a salon from a list below.",
            registerRoleLabel: "I am registering as",
            roleCustomer: "Customer",
            roleEmployee: "Employee",
            roleOwner: "Owner",
            localeLabel: "Language",
            languageEnglish: "English",
            languageSerbian: "Serbian",
            backToFejd: "Back to Fejd",
            username: "Username",
            email: "Email",
            firstName: "First name",
            lastName: "Last name",
            password: "Password",
            passwordConfirm: "Confirm password",
            rememberMe: "Remember me",
            doLogIn: "Log in",
            doRegister: "Register",
            doSubmit: "Submit",
            backToLogin: "Back to login",
            emailInstruction: "Enter your email or username and we'll send you a link to reset your password.",
            emailInstructionUsername: "Enter your username and we'll send you a link to reset your password.",
        },
        sr: {
            loginTitle: "Prijava na Fejd",
            registerTitle: "Napravite svoj Fejd nalog",
            forgotPasswordTitle: "Resetovanje lozinke",
            loginIntro: "Prijavite se da biste zakazali termine i upravljali svojim salonom.",
            usernameOrEmail: "Email ili korisničko ime",
            forgotPassword: "Zaboravili ste lozinku?",
            alreadyHaveAccount: "Već imate nalog?",
            noAccount: "Nemate nalog?",
            homeIntro: "Zakažite termine šišanja. Otvorite salon putem liste ispod.",
            registerRoleLabel: "Registrujem se kao",
            roleCustomer: "Klijent",
            roleEmployee: "Zaposleni",
            roleOwner: "Vlasnik",
            localeLabel: "Jezik",
            languageEnglish: "Engleski",
            languageSerbian: "Srpski",
            backToFejd: "Nazad na Fejd",
            username: "Korisničko ime",
            email: "Email",
            firstName: "Ime",
            lastName: "Prezime",
            password: "Lozinka",
            passwordConfirm: "Potvrdite lozinku",
            rememberMe: "Zapamti me",
            doLogIn: "Prijavi se",
            doRegister: "Registruj se",
            doSubmit: "Pošalji",
            backToLogin: "Nazad na prijavu",
            emailInstruction: "Unesite email adresu ili korisničko ime i poslaćemo vam link za resetovanje lozinke.",
            emailInstructionUsername: "Unesite korisničko ime i poslaćemo vam link za resetovanje lozinke.",
        },
    })
    .build();

type I18n = typeof ofTypeI18n;

export { useI18n, type I18n };
