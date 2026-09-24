import type { ExtendKcContext } from "keycloakify/login";

export type KcContextExtension = {
    properties?: Record<string, string | undefined>;
};

export type KcContextExtensionPerPage = {};

export type KcContext = ExtendKcContext<KcContextExtension, KcContextExtensionPerPage>;
