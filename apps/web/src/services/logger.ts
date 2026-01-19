type LogLevel = 'info' | 'warn' | 'error' | 'debug';

class LoggerService {
    private static instance: LoggerService;

    private constructor() { }

    public static getInstance(): LoggerService {
        if (!LoggerService.instance) {
            LoggerService.instance = new LoggerService();
        }
        return LoggerService.instance;
    }

    private log(level: LogLevel, message: string, ...args: unknown[]) {
        // In production, this would send data to Datadog/Sentry
        const timestamp = new Date().toISOString();
        const prefix = `[${timestamp}] [${level.toUpperCase()}]`;

        switch (level) {
            case 'info':
                console.info(prefix, message, ...args);
                break;
            case 'warn':
                console.warn(prefix, message, ...args);
                break;
            case 'error':
                console.error(prefix, message, ...args);
                break;
            case 'debug':
                if (process.env.NODE_ENV === 'development') {
                    console.debug(prefix, message, ...args);
                }
                break;
        }
    }

    public info(message: string, ...args: unknown[]) {
        this.log('info', message, ...args);
    }

    public warn(message: string, ...args: unknown[]) {
        this.log('warn', message, ...args);
    }

    public error(message: string, ...args: unknown[]) {
        this.log('error', message, ...args);
    }

    public debug(message: string, ...args: unknown[]) {
        this.log('debug', message, ...args);
    }
}

export const logger = LoggerService.getInstance();
