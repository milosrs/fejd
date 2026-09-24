var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g = Object.create((typeof Iterator === "function" ? Iterator : Object).prototype);
    return g.next = verb(0), g["throw"] = verb(1), g["return"] = verb(2), typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (g && (g = 0, op[0] && (_ = 0)), _) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { keycloakify } from 'keycloakify/vite-plugin';
import path from 'node:path';
import fs from 'node:fs';
// The theme is built from React (this project) and keycloakify emits the FTL
// theme directly into the repo's mounted theme directory: keycloak/themes/fejd.
export default defineConfig({
    plugins: [
        react(),
        keycloakify({
            accountThemeImplementation: 'none',
            themeName: 'fejd',
            // Absolute URL of the Fejd app; the nav bar links back to it. Resolved at
            // runtime from the FEJD_APP_URL environment variable (see docker-compose).
            extraThemeProperties: ['fejdAppUrl=${env.FEJD_APP_URL:https://fejd.fyi}'],
            keycloakifyBuildDirPath: path.resolve(process.cwd(), '../themes'),
            // keycloakify requires at least one JAR target; keep a single one (the
            // repo mounts the theme directory directly, so the JAR is only a build
            // by-product and is gitignored).
            keycloakVersionTargets: {
                '22-to-25': false,
                'all-other-versions': true,
            },
            // keycloakify stages the generated theme under
            // <keycloakifyBuildDirPath>/resources/theme/<themeName> and deletes it
            // after packaging the JARs. Copy it out to the mounted theme directory
            // before that cleanup runs.
            postBuild: function (buildContext) { return __awaiter(void 0, void 0, void 0, function () {
                var themeName, srcDir, destDir;
                return __generator(this, function (_a) {
                    themeName = buildContext.themeNames[0];
                    srcDir = path.join(process.cwd(), 'theme', themeName);
                    destDir = path.join(buildContext.keycloakifyBuildDirPath, themeName);
                    fs.rmSync(destDir, { recursive: true, force: true });
                    fs.cpSync(srcDir, destDir, { recursive: true });
                    return [2 /*return*/];
                });
            }); },
        }),
    ],
});
