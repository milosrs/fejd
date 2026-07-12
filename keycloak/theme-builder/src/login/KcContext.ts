export type KcContext = {
  pageId: string;
  url: {
    loginAction: string;
    resourcesPath: string;
    themePath: string;
  };
  message?: {
    summary?: string;
  };
  login?: {
    username?: string;
  };
};
