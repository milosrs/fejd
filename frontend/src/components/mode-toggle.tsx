import { Moon, Sun } from "lucide-react"
import { Menu, MenuItem, MenuTrigger, Popover } from "react-aria-components"

import { Button } from "#components/ui/button"
import { useTheme } from "#components/theme-provider"

const menuItemClassName =
    "flex w-full cursor-default items-center gap-2 rounded-lg px-2 py-1.5 text-sm outline-none select-none data-focused:bg-accent data-focused:text-accent-foreground"

export function ModeToggle() {
    const { setTheme } = useTheme()

    return (
        <MenuTrigger>
            <Button variant="outline" size="icon" aria-label="Toggle theme">
                <Sun className="size-5 scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
                <Moon className="absolute size-5 scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0" />
            </Button>
            <Popover className="min-w-32 rounded-xl bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-foreground/5 dark:ring-foreground/10">
                <Menu
                    className="outline-none"
                    onAction={(key) => setTheme(key as "light" | "dark" | "system")}
                >
                    <MenuItem id="light" className={menuItemClassName}>
                        Light
                    </MenuItem>
                    <MenuItem id="dark" className={menuItemClassName}>
                        Dark
                    </MenuItem>
                    <MenuItem id="system" className={menuItemClassName}>
                        System
                    </MenuItem>
                </Menu>
            </Popover>
        </MenuTrigger>
    )
}
