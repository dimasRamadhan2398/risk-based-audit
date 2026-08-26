import { calendar } from "#build/ui";

export default defineAppConfig({
  /**
   * Icon system (Nuxt UI + Iconify)
   */
  icon: {
    provider: "iconify",
    aliases: {
      // basic UI
      close: "heroicons:x-mark",
      clock: "heroicons:clock",
      check: "heroicons:check",
      loading: "heroicons:arrow-path",
      charter: "heroicons:clipboard-document-list",
      warning: "heroicons:exclamation-triangle",
      alert: "heroicons:alert-circle",
      calendar: "heroicons:calendar",
      download: "heroicons:arrow-down-tray",
      pin: "heroicons:map-pin",
      // navigation
      chevronDown: "heroicons:chevron-down",
      chevronRight: "heroicons:chevron-right",
      chevronUp: "heroicons:chevron-up",
      logout: "lucide:log-out",
      setting: "lucide:settings",
      language: "lucide:languages",

      // actions
      add: "heroicons:plus",
      edit: "heroicons:pencil",
      delete: "heroicons:trash",
      import: "heroicons:document-arrow-up",
      plan: "lucide:target",
      info: "heroicons:information-circle",
      copy: "heroicons:document-duplicate"
    },
  },
  theme: {
    dark: false,
    primaryColor: "primary",
  },

  ui: {
    accordion: {
      slots: {
        root: 'w-full gap-4 flex flex-col',
        item: 'border-all border-2 p-2 border-neutral-500 last:border-b-2 rounded-xl pl-4',
        header: 'flex',
        trigger: 'group flex-1 flex items-center gap-1.5 font-medium text-sm py-3.5 focus-visible:outline-primary min-w-0',
        content: 'data-[state=open]:animate-[accordion-down_200ms_ease-out] data-[state=closed]:animate-[accordion-up_200ms_ease-out] overflow-hidden focus:outline-none p-2',
        body: 'text-sm',
        leadingIcon: 'shrink-0 size-5 text-secondary',
        trailingIcon: 'shrink-0 size-5 ms-auto group-data-[state=open]:rotate-180 transition-transform duration-200',
        label: 'text-start wrap-break-word'
      },
      variants: {
        disabled: {
          true: {
            trigger: 'cursor-not-allowed opacity-75'
          }
        }
      }
    },
    alert: {
  slots: {
    // pl-5 sengaja lebih lebar dari pr-4 untuk kasih ruang si "spine" tab di kiri
    root: 'relative isolate overflow-hidden w-full rounded-lg pl-5 pr-4 py-3.5 flex gap-3 ring-1 ring-transparent transition-colors duration-200 before:absolute before:inset-y-2.5 before:left-2 before:w-[3px] before:rounded-full before:content-[""] before:transition-colors before:duration-200',
    wrapper: 'min-w-0 flex-1 flex flex-col gap-0.5',
    title: 'font-semibold text-[15px] tracking-tight leading-snug',
    description: 'text-[13px] leading-relaxed opacity-80 font-normal',
    icon: 'shrink-0 size-5 mt-0.5 opacity-90',
    avatar: 'shrink-0',
    avatarSize: '2xl',
    actions: 'flex flex-wrap gap-2 shrink-0 mt-2',
    close: 'p-0 -mr-1 -mt-1'
  },
  variants: {
    color: {
      primary: '',
      secondary: '',
      success: '',
      info: '',
      warning: '',
      error: '',
      neutral: ''
    },
    variant: {
      solid: '',
      outline: '',
      soft: '',
      subtle: '',
      none: ''
    },
    orientation: {
      horizontal: {
        root: 'items-center',
        actions: 'items-center'
      },
      vertical: {
        root: 'items-start',
        actions: 'items-start mt-2.5'
      }
    },
    title: {
      true: {
        description: 'mt-1'
      }
    }
  },
  compoundVariants: [
    // ===== Primary =====
    { color: 'primary', variant: 'solid', class: { root: 'bg-primary-600 dark:bg-primary-500 text-white before:bg-primary-300 dark:before:bg-primary-200' } },
    { color: 'primary', variant: 'outline', class: { root: 'bg-white dark:bg-neutral-900 ring-primary-500/30 dark:ring-primary-400/30 text-primary-700 dark:text-primary-400 before:bg-primary-500 dark:before:bg-primary-400' } },
    { color: 'primary', variant: 'soft', class: { root: 'bg-primary-500/10 dark:bg-primary-400/10 text-primary-700 dark:text-primary-400 before:bg-primary-500 dark:before:bg-primary-400' } },
    { color: 'primary', variant: 'subtle', class: { root: 'bg-primary-500/10 dark:bg-primary-400/10 ring-primary-500/25 dark:ring-primary-400/25 text-primary-700 dark:text-primary-400 before:bg-primary-500 dark:before:bg-primary-400' } },
    { color: 'primary', variant: 'none', class: { root: 'bg-transparent ring-0 p-0 pl-0 before:hidden text-primary-700 dark:text-primary-400' } },

    // ===== Secondary =====
    { color: 'secondary', variant: 'solid', class: { root: 'bg-secondary-600 dark:bg-secondary-500 text-white before:bg-secondary-300 dark:before:bg-secondary-200' } },
    { color: 'secondary', variant: 'outline', class: { root: 'bg-white dark:bg-neutral-900 ring-secondary-500/30 dark:ring-secondary-400/30 text-secondary-700 dark:text-secondary-400 before:bg-secondary-500 dark:before:bg-secondary-400' } },
    { color: 'secondary', variant: 'soft', class: { root: 'bg-secondary-500/10 dark:bg-secondary-400/10 text-secondary-700 dark:text-secondary-400 before:bg-secondary-500 dark:before:bg-secondary-400' } },
    { color: 'secondary', variant: 'subtle', class: { root: 'bg-secondary-500/10 dark:bg-secondary-400/10 ring-secondary-500/25 dark:ring-secondary-400/25 text-secondary-700 dark:text-secondary-400 before:bg-secondary-500 dark:before:bg-secondary-400' } },
    { color: 'secondary', variant: 'none', class: { root: 'bg-transparent ring-0 p-0 pl-0 before:hidden text-secondary-700 dark:text-secondary-400' } },

    // ===== Success =====
    { color: 'success', variant: 'solid', class: { root: 'bg-success-600 dark:bg-success-500 text-white before:bg-success-300 dark:before:bg-success-200' } },
    { color: 'success', variant: 'outline', class: { root: 'bg-white dark:bg-neutral-900 ring-success-500/30 dark:ring-success-400/30 text-success-700 dark:text-success-400 before:bg-success-500 dark:before:bg-success-400' } },
    { color: 'success', variant: 'soft', class: { root: 'bg-success-500/10 dark:bg-success-400/10 text-success-700 dark:text-success-400 before:bg-success-500 dark:before:bg-success-400' } },
    { color: 'success', variant: 'subtle', class: { root: 'bg-success-500/10 dark:bg-success-400/10 ring-success-500/25 dark:ring-success-400/25 text-success-700 dark:text-success-400 before:bg-success-500 dark:before:bg-success-400' } },
    { color: 'success', variant: 'none', class: { root: 'bg-transparent ring-0 p-0 pl-0 before:hidden text-success-700 dark:text-success-400' } },

    // ===== Info =====
    { color: 'info', variant: 'solid', class: { root: 'bg-info-600 dark:bg-info-500 text-white before:bg-info-300 dark:before:bg-info-200' } },
    { color: 'info', variant: 'outline', class: { root: 'bg-white dark:bg-neutral-900 ring-info-500/30 dark:ring-info-400/30 text-info-700 dark:text-info-400 before:bg-info-500 dark:before:bg-info-400' } },
    { color: 'info', variant: 'soft', class: { root: 'bg-info-500/10 dark:bg-info-400/10 text-info-700 dark:text-info-400 before:bg-info-500 dark:before:bg-info-400' } },
    { color: 'info', variant: 'subtle', class: { root: 'bg-info-500/10 dark:bg-info-400/10 ring-info-500/25 dark:ring-info-400/25 text-info-700 dark:text-info-400 before:bg-info-500 dark:before:bg-info-400' } },
    { color: 'info', variant: 'none', class: { root: 'bg-transparent ring-0 p-0 pl-0 before:hidden text-info-700 dark:text-info-400' } },

    // ===== Warning (fill terang -> teks gelap, spine pakai gelap transparan biar tetap ada kontras "stempel") =====
    { color: 'warning', variant: 'solid', class: { root: 'bg-warning-500 dark:bg-warning-500 text-neutral-950 font-medium before:bg-neutral-950/25' } },
    { color: 'warning', variant: 'outline', class: { root: 'bg-white dark:bg-neutral-900 ring-warning-500/30 dark:ring-warning-400/30 text-warning-700 dark:text-warning-400 before:bg-warning-500 dark:before:bg-warning-400' } },
    { color: 'warning', variant: 'soft', class: { root: 'bg-warning-500/10 dark:bg-warning-400/10 text-warning-700 dark:text-warning-400 before:bg-warning-500 dark:before:bg-warning-400' } },
    { color: 'warning', variant: 'subtle', class: { root: 'bg-warning-500/10 dark:bg-warning-400/10 ring-warning-500/25 dark:ring-warning-400/25 text-warning-700 dark:text-warning-400 before:bg-warning-500 dark:before:bg-warning-400' } },
    { color: 'warning', variant: 'none', class: { root: 'bg-transparent ring-0 p-0 pl-0 before:hidden text-warning-700 dark:text-warning-400' } },

    // ===== Error =====
    { color: 'error', variant: 'solid', class: { root: 'bg-error-600 dark:bg-error-500 text-white before:bg-error-300 dark:before:bg-error-200' } },
    { color: 'error', variant: 'outline', class: { root: 'bg-white dark:bg-neutral-900 ring-error-500/30 dark:ring-error-400/30 text-error-700 dark:text-error-400 before:bg-error-500 dark:before:bg-error-400' } },
    { color: 'error', variant: 'soft', class: { root: 'bg-error-500/10 dark:bg-error-400/10 text-error-700 dark:text-error-400 before:bg-error-500 dark:before:bg-error-400' } },
    { color: 'error', variant: 'subtle', class: { root: 'bg-error-500/10 dark:bg-error-400/10 ring-error-500/25 dark:ring-error-400/25 text-error-700 dark:text-error-400 before:bg-error-500 dark:before:bg-error-400' } },
    { color: 'error', variant: 'none', class: { root: 'bg-transparent ring-0 p-0 pl-0 before:hidden text-error-700 dark:text-error-400' } },

    // ===== Neutral (inverted di solid, biar tetap punya "berat" tanpa nyolok) =====
    { color: 'neutral', variant: 'solid', class: { root: 'bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-900 before:bg-neutral-500 dark:before:bg-neutral-400' } },
    { color: 'neutral', variant: 'outline', class: { root: 'bg-white dark:bg-neutral-900 ring-neutral-300 dark:ring-neutral-700 text-neutral-800 dark:text-neutral-200 before:bg-neutral-400 dark:before:bg-neutral-500' } },
    { color: 'neutral', variant: 'soft', class: { root: 'bg-neutral-100 dark:bg-neutral-800/60 text-neutral-800 dark:text-neutral-200 before:bg-neutral-400 dark:before:bg-neutral-500' } },
    { color: 'neutral', variant: 'subtle', class: { root: 'bg-neutral-100 dark:bg-neutral-800/60 ring-neutral-300 dark:ring-neutral-700 text-neutral-800 dark:text-neutral-200 before:bg-neutral-400 dark:before:bg-neutral-500' } },
    { color: 'neutral', variant: 'none', class: { root: 'bg-transparent ring-0 p-0 pl-0 before:hidden text-neutral-800 dark:text-neutral-200' } }
  ],
  defaultVariants: {
    color: 'primary',
    variant: 'solid'
  }
},
    badge: {
      slots: {
        base: "font-medium inline-flex items-center",
        label: "truncate font-semibold py-1 px-2",
        leadingIcon: "shrink-0",
        leadingAvatar: "shrink-0",
        leadingAvatarSize: "",
        trailingIcon: "shrink-0",
      },
      variants: {
        fieldGroup: {
          horizontal:
            "not-only:first:rounded-e-none not-only:last:rounded-s-none not-last:not-first:rounded-none focus-visible:z-[1]",
          vertical:
            "not-only:first:rounded-b-none not-only:last:rounded-t-none not-last:not-first:rounded-none focus-visible:z-[1]",
        },
        color: {
          successLight: "success",
          successDark: "success",
          secondary: "secondary",
          success: "success",
          info: "info",
          warningLight: "warning",
          warningDark: "warning",
          error: "error",
          neutral: "neutral",
        },
        variant: {
          solid: "",
          outline: "",
          soft: "bg-primary-200 text-primary-900",
          subtle: "",
        },
        size: {
          xs: {
            base: "text-[8px]/3 px-1 py-0.5 gap-1 rounded-sm",
            leadingIcon: "size-3",
            leadingAvatarSize: "3xs",
            trailingIcon: "size-3",
          },
          sm: {
            base: "text-[10px]/3 px-1.5 py-1 gap-1 rounded-sm",
            leadingIcon: "size-3",
            leadingAvatarSize: "3xs",
            trailingIcon: "size-3",
          },
          md: {
            base: "text-xs px-2 py-1 gap-1 rounded-md",
            leadingIcon: "size-4",
            leadingAvatarSize: "3xs",
            trailingIcon: "size-4",
          },
          lg: {
            base: "text-sm px-2 py-1 gap-1.5 rounded-md",
            leadingIcon: "size-5",
            leadingAvatarSize: "2xs",
            trailingIcon: "size-5",
          },
          xl: {
            base: "text-base px-2.5 py-1 gap-1.5 rounded-md",
            leadingIcon: "size-6",
            leadingAvatarSize: "2xs",
            trailingIcon: "size-6",
          },
        },
        square: {
          true: "",
        },
      },
      compoundVariants: [
        {
          color: "primary",
          variant: "solid",
          class: "bg-primary text-inverted",
        },
        {
          color: "primary",
          variant: "outline",
          class: "text-primary ring ring-inset ring-primary/50",
        },
        {
          color: "primary",
          variant: "soft",
          class: "bg-primary-500/50 text-primary-600",
        },
        {
          color: "primary",
          variant: "subtle",
          class: "bg-primary/10 text-primary ring ring-inset ring-primary/25",
        },
        {
          color: 'success',
          variant: 'solid',
          class: "bg-success-600 text-highlighted",
        },
        {
          color: 'success',
          variant: 'subtle',
          class: "text-success-600 ring-success/100 bg-success-100",
        },
        {
          color: 'success',
          variant: 'soft',
          class: "text-success-500 ring-success/100 bg-success-200/20 border-none rounded-xl",
        },
        {
          color: "successDark",
          variant: "solid",
          class: "bg-success-800 text-inverted",
        },
        {
          color: "successDark",
          variant: "outline",
          class: "text-success-800 ring ring-inset ring-success-800/50",
        },
        {
          color: "successDark",
          variant: "soft",
          class: "bg-success-900/20 text-success-800",
        },
        {
          color: "successDark",
          variant: "subtle",
          class:
            "bg-success-800/10 text-success-800 ring ring-inset ring-success-800/25",
        },
        {
          color: "successLight",
          variant: "solid",
          class: "bg-success-400 text-inverted",
        },
        {
          color: "successLight",
          variant: "outline",
          class: "text-success-400 ring ring-inset ring-success-400/50",
        },
        {
          color: "successLight",
          variant: "soft",
          class: "bg-success-400/20 text-success-600",
        },
        {
          color: "successLight",
          variant: "subtle",
          class:
            "bg-success-400/10 text-success-400 ring ring-inset ring-success-400/25",
        },
        {
          color: "warningDark",
          variant: "solid",
          class: "bg-primary-200 text-inverted",
        },
        {
          color: "warningDark",
          variant: "outline",
          class: "text-primary-200 ring ring-inset ring-primary-200/50",
        },
        {
          color: "warningDark",
          variant: "soft",
          class: "bg-primary-200/20 text-primary-600",
        },
        {
          color: "warningDark",
          variant: "subtle",
          class:
            "bg-primary-200/10 text-primary-200 ring ring-inset ring-primary-200/25",
        },
        {
          color: "warningLight",
          variant: "solid",
          class: "bg-warning-400 text-inverted",
        },
        {
          color: "warningLight",
          variant: "outline",
          class: "text-warning-400 ring ring-inset ring-warning-400/50",
        },
        {
          color: "warningLight",
          variant: "soft",
          class: "bg-warning-400/20 text-warning-600",
        },
        {
          color: "warningLight",
          variant: "subtle",
          class:
            "bg-warning-400/10 text-warning-400 ring ring-inset ring-warning-400/25",
        },
        {
          color: "neutral",
          variant: "solid",
          class: "text-inverted bg-inverted",
        },
        {
          color: "neutral",
          variant: "outline",
          class: "ring ring-inset ring-accented text-default bg-default",
        },
        {
          color: "neutral",
          variant: "soft",
          class: "text-default bg-elevated",
        },
        {
          color: "neutral",
          variant: "subtle",
          class: "ring ring-inset ring-accented text-default bg-elevated",
        },
        {
          size: "xs",
          square: true,
          class: "p-0.5",
        },
        {
          size: "sm",
          square: true,
          class: "p-1",
        },
        {
          size: "md",
          square: true,
          class: "p-1",
        },
        {
          size: "lg",
          square: true,
          class: "p-1",
        },
        {
          size: "xl",
          square: true,
          class: "p-1",
        },
      ],
      defaultVariants: {
        color: "primary",
        variant: "solid",
        size: "md",
      },
    },
    checkbox: {
      slots: {
        root: "relative flex items-start",
        container: "flex items-center",
        base: "rounded-sm ring ring-inset ring-accented overflow-hidden focus-visible:outline-2 focus-visible:outline-offset-2",
        indicator: "flex items-center justify-center size-full text-inverted",
        icon: "shrink-0 size-full",
        wrapper: "w-full",
        label: "block",
        description: "text-muted",
      },
      variants: {
        color: {
          primary: {
            base: "focus-visible:outline-primary",
            indicator: "bg-primary",
          },
          secondary: {
            base: "focus-visible:outline-secondary",
            indicator: "bg-secondary",
          },
          success: {
            base: "focus-visible:outline-success",
            indicator: "bg-success",
          },
          info: {
            base: "focus-visible:outline-info",
            indicator: "bg-info",
          },
          warning: {
            base: "focus-visible:outline-warning",
            indicator: "bg-warning",
          },
          error: {
            base: "focus-visible:outline-error",
            indicator: "bg-error",
          },
          neutral: {
            base: "focus-visible:outline-inverted",
            indicator: "bg-inverted",
          },
        },
        variant: {
          list: {
            root: "",
          },
          card: {
            root: "border border-muted rounded-lg",
          },
        },
        indicator: {
          start: {
            root: "flex-row",
            wrapper: "ms-2",
          },
          end: {
            root: "flex-row-reverse",
            wrapper: "me-2",
          },
          hidden: {
            base: "sr-only",
            wrapper: "text-center",
          },
        },
        size: {
          xs: {
            base: "size-3",
            container: "h-4",
            wrapper: "text-xs",
          },
          sm: {
            base: "size-3.5",
            container: "h-4",
            wrapper: "text-xs",
          },
          md: {
            base: "size-4",
            container: "h-5",
            wrapper: "text-sm",
          },
          lg: {
            base: "size-4.5",
            container: "h-5",
            wrapper: "text-sm",
          },
          xl: {
            base: "size-5",
            container: "h-6",
            wrapper: "text-base",
          },
        },
        required: {
          true: {
            label: "after:content-['*'] after:ms-0.5 after:text-error",
          },
        },
        disabled: {
          true: {
            root: "opacity-75",
            base: "cursor-not-allowed",
            label: "cursor-not-allowed",
            description: "cursor-not-allowed",
          },
        },
        checked: {
          true: "",
        },
      },
      compoundVariants: [
        {
          size: "xs",
          variant: "card",
          class: {
            root: "p-2.5",
          },
        },
        {
          size: "sm",
          variant: "card",
          class: {
            root: "p-3",
          },
        },
        {
          size: "md",
          variant: "card",
          class: {
            root: "p-3.5",
          },
        },
        {
          size: "lg",
          variant: "card",
          class: {
            root: "p-4",
          },
        },
        {
          size: "xl",
          variant: "card",
          class: {
            root: "p-4.5",
          },
        },
        {
          color: "primary",
          variant: "card",
          class: {
            root: "has-data-[state=checked]:border-primary",
          },
        },
        {
          color: "neutral",
          variant: "card",
          class: {
            root: "has-data-[state=checked]:border-inverted",
          },
        },
        {
          variant: "card",
          disabled: true,
          class: {
            root: "cursor-not-allowed",
          },
        },
      ],
      defaultVariants: {
        size: "md",
        color: "primary",
        variant: "list",
        indicator: "start",
      },
    },

    container: {
      base: "max-w-(--ui-container) w-full space-y-8 bg-[var(--bg-surface)] px-8 py-4 rounded-xl z-100 block relative shadow-xl border border-[var(--border-main)] transition-all duration-300",
    },
    card: {
      slots: {
        root: "rounded-xl overflow-hidden border-2 border-slate-300 transition-all duration-300 shadow-md hover:shadow-lg",
        header: "p-4 sm:px-6 border-b border-[var(--border-main)] bg-[var(--bg-surface)]",
        body: "p-4 sm:p-6",
        footer: "p-4 sm:px-6 border-t border-[var(--border-main)] bg-[var(--bg-surface)]",
      },
      variants: {
        variant: {
          solid: {
            root: "bg-[var(--bg-main)] text-[var(--text-main)]",
          },
          outline: {
            root: "bg-[var(--bg-main)] border border-[var(--border-main)]",
          },
          soft: {
            root: "bg-[var(--bg-surface)] rounded-xl z-100 block relative shadow-lg",
          },
          subtle: {
            root: "bg-[var(--bg-surface)]/50 ring ring-[var(--border-main)] divide-y divide-[var(--border-main)]",
          },
        },
      },
      defaultVariants: {
        variant: "outline",
      },
    },
    colors: {
      primary: "primary",
      secondary: "secondary",
      success: "success",
      info: "info",
      warning: "warning",
      error: "error",
      neutral: "neutral",
    },
    formField: {
      slots: {
        root: "",
        wrapper: '',
        labelWrapper:
          "flex flex-shrink-1 content-start items-start justify-between",
        label: "block font-medium text-[var(--text-main)] min-w-24",
        container: "relative flex-1 min-w-0",
        description: "text-[var(--text-muted)] ",
        error: "mt-2 text-error-500 dark:text-error-400",
        hint: "text-[var(--text-muted)]",
        help: "mt-2 text-[var(--text-muted)]",
      },
      variants: {
        size: {
          xs: {
            root: "text-xs",
          },
          sm: {
            root: "text-xs",
          },
          md: {
            root: "text-sm",
          },
          lg: {
            root: "text-sm",
          },
          xl: {
            root: "text-base",
          },
        },
        required: {
          true: {
            label: "after:content-['*'] after:ms-0.5 after:text-error",
          },
        },
        orientation: {
          vertical: {
            container: "mt-1",
          },
          horizontal: {
            root: "flex flex-row justify-between place-items-baseline",
          },
        },
      },
      defaultVariants: {
        size: "md",
        orientation: "vertical",
      },
    },
    input: {
      slots: {
        root: 'relative inline-flex items-center w-full flex flex-1',
        base: [
          'w-full rounded-md border-0 appearance-none placeholder:text-dimmed focus:outline-none disabled:cursor-not-allowed disabled:opacity-75',
          'transition-colors'
        ],
        leading: 'absolute inset-y-0 start-0 flex items-center',
        leadingIcon: 'shrink-0 text-dimmed',
        leadingAvatar: 'shrink-0',
        leadingAvatarSize: '',
        trailing: 'absolute inset-y-0 end-0 flex items-center',
        trailingIcon: 'shrink-0 text-dimmed'
      },
      variants: {
        fieldGroup: {
          horizontal: {
            root: 'group has-focus-visible:z-[1]',
            base: 'group-not-only:group-first:rounded-e-none group-not-only:group-last:rounded-s-none group-not-last:group-not-first:rounded-none'
          },
          vertical: {
            root: 'group has-focus-visible:z-[1]',
            base: 'group-not-only:group-first:rounded-b-none group-not-only:group-last:rounded-t-none group-not-last:group-not-first:rounded-none'
          }
        },
        size: {
          xs: {
            base: 'px-2 py-1 text-sm/4 gap-1',
            leading: 'ps-2',
            trailing: 'pe-2',
            leadingIcon: 'size-4',
            leadingAvatarSize: '3xs',
            trailingIcon: 'size-4'
          },
          sm: {
            base: 'px-2.5 py-1.5 text-sm/4 gap-1.5',
            leading: 'ps-2.5',
            trailing: 'pe-2.5',
            leadingIcon: 'size-4',
            leadingAvatarSize: '3xs',
            trailingIcon: 'size-4'
          },
          md: {
            base: 'px-2.5 py-1.5 text-base/5 gap-1.5',
            leading: 'ps-2.5',
            trailing: 'pe-2.5',
            leadingIcon: 'size-5',
            leadingAvatarSize: '2xs',
            trailingIcon: 'size-5'
          },
          lg: {
            base: 'px-3 py-2 text-base/5 gap-2',
            leading: 'ps-3',
            trailing: 'pe-3',
            leadingIcon: 'size-5',
            leadingAvatarSize: '2xs',
            trailingIcon: 'size-5'
          },
          xl: {
            base: 'px-3 py-2 text-base gap-2',
            leading: 'ps-3',
            trailing: 'pe-3',
            leadingIcon: 'size-6',
            leadingAvatarSize: 'xs',
            trailingIcon: 'size-6'
          }
        },
        variant: {
          outline: 'text-highlighted bg-default ring ring-inset ring-accented',
          soft: 'text-highlighted bg-elevated/50 hover:bg-elevated focus:bg-elevated disabled:bg-elevated/50',
          subtle: 'text-highlighted bg-elevated ring ring-inset ring-accented',
          ghost: 'text-highlighted bg-transparent hover:bg-elevated focus:bg-elevated disabled:bg-transparent dark:disabled:bg-transparent',
          none: 'text-highlighted bg-transparent'
        },
        color: {
          primary: '',
          secondary: '',
          success: '',
          info: '',
          warning: '',
          error: '',
          neutral: ''
        },
        leading: {
          true: ''
        },
        trailing: {
          true: ''
        },
        loading: {
          true: ''
        },
        highlight: {
          true: ''
        },
        fixed: {
          false: ''
        },
        type: {
          file: 'file:me-1.5 file:font-medium file:text-muted file:outline-none'
        }
      },
      compoundVariants: [
        {
          color: 'primary',
          variant: [
            'outline',
            'subtle'
          ],
          class: 'focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary'
        },
        {
          color: 'primary',
          highlight: true,
          class: 'ring ring-inset ring-primary'
        },
        {
          color: 'neutral',
          variant: [
            'outline',
            'subtle'
          ],
          class: 'focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-inverted'
        },
        {
          color: 'neutral',
          highlight: true,
          class: 'ring ring-inset ring-inverted'
        },
        {
          leading: true,
          size: 'xs',
          class: 'ps-7'
        },
        {
          leading: true,
          size: 'sm',
          class: 'ps-8'
        },
        {
          leading: true,
          size: 'md',
          class: 'ps-9'
        },
        {
          leading: true,
          size: 'lg',
          class: 'ps-10'
        },
        {
          leading: true,
          size: 'xl',
          class: 'ps-11'
        },
        {
          trailing: true,
          size: 'xs',
          class: 'pe-7'
        },
        {
          trailing: true,
          size: 'sm',
          class: 'pe-8'
        },
        {
          trailing: true,
          size: 'md',
          class: 'pe-9'
        },
        {
          trailing: true,
          size: 'lg',
          class: 'pe-10'
        },
        {
          trailing: true,
          size: 'xl',
          class: 'pe-11'
        },
        {
          loading: true,
          leading: true,
          class: {
            leadingIcon: 'animate-spin'
          }
        },
        {
          loading: true,
          leading: false,
          trailing: true,
          class: {
            trailingIcon: 'animate-spin'
          }
        },
        {
          fixed: false,
          size: 'xs',
          class: 'md:text-xs'
        },
        {
          fixed: false,
          size: 'sm',
          class: 'md:text-xs'
        },
        {
          fixed: false,
          size: 'md',
          class: 'md:text-sm'
        },
        {
          fixed: false,
          size: 'lg',
          class: 'md:text-sm'
        }
      ],
      defaultVariants: {
        size: 'md',
        color: 'primary',
        variant: 'outline'
      }
    },
    main: {
      base: "min-h-[calc(100vh-var(--ui-header-height))] max-w-8xl px-24 py-16",
    },
    modal: {
      slots: {
        overlay: "fixed inset-0",
        content:
          "bg-[var(--bg-surface)] text-[var(--text-main)] border border-[var(--border-main)] divide-y divide-[var(--border-main)] flex flex-col focus:outline-none",
        header: "flex items-center gap-1.5 p-4 sm:px-6 min-h-16 text-[var(--text-main)]",
        wrapper: "",
        body: "flex-1 p-4 sm:p-6 text-[var(--text-main)]",
        footer: "flex flex-row justify-between gap-1.5 p-4 sm:px-6 text-[var(--text-main)]",
        title: "text-[var(--text-main)] font-semibold text-2xl",
        description: "mt-1 text-[var(--text-muted)] text-sm",
        close: "absolute top-4 end-4 text-[var(--text-muted)] hover:text-[var(--text-main)]",
      },
      variants: {
        transition: {
          true: {
            overlay:
              "data-[state=open]:animate-[fade-in_200ms_ease-out] data-[state=closed]:animate-[fade-out_200ms_ease-in]",
            content:
              "data-[state=open]:animate-[scale-in_200ms_ease-out] data-[state=closed]:animate-[scale-out_200ms_ease-in]",
          },
        },
        fullscreen: {
          true: {
            content: "inset-0",
          },
          false: {
            content:
              "w-[calc(100vw-1rem)] min-w-4xl max-w-lg rounded-lg shadow-lg ring ring-default",
          },
        },
        overlay: {
          true: {
            overlay: "bg-elevated/75",
          },
        },
        scrollable: {
          true: {
            overlay: "overflow-y-auto",
            content: "relative",
          },
          false: {
            content: "fixed",
            body: "overflow-y-auto",
          },
        },
      },
      compoundVariants: [
        {
          scrollable: true,
          fullscreen: false,
          class: {
            overlay: "grid place-items-center p-4 sm:py-8",
          },
        },
        {
          scrollable: false,
          fullscreen: false,
          class: {
            content:
              "top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 max-h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-4rem)] overflow-hidden",
          },
        },
      ],
    },
    navigationMenu: {
      slots: {
        root: 'relative flex gap-1.5 [&>div]:min-w-0',
        list: 'isolate min-w-0',
        label: 'w-full flex items-center gap-1.5 font-semibold text-xs/5 text-highlighted px-2.5 py-1.5',
        item: 'min-w-0 py-1',
        link: 'group relative w-full flex items-center gap-1.5 font-medium text-sm before:absolute before:z-[-1] before:rounded-md focus:outline-none focus-visible:outline-none dark:focus-visible:outline-none focus-visible:before:ring-inset focus-visible:before:ring-2 rounded-xl',
        linkLeadingIcon: 'shrink-0 size-5',
        linkLeadingAvatar: 'shrink-0',
        linkLeadingAvatarSize: '2xs',
        linkLeadingChipSize: 'sm',
        linkTrailing: 'group ms-auto inline-flex gap-1.5 items-center',
        linkTrailingBadge: 'shrink-0',
        linkTrailingBadgeSize: 'sm',
        linkTrailingIcon: 'size-5 transform shrink-0 group-data-[state=open]:rotate-180 transition-transform duration-200',
        linkLabel: 'truncate',
        linkLabelExternalIcon: 'inline-block size-3 align-top text-dimmed',
        childList: 'isolate',
        childLabel: 'text-xs text-highlighted',
        childItem: '',
        childLink: 'group relative size-full flex items-start text-start text-sm before:absolute before:z-[-1] before:rounded-md focus:outline-none focus-visible:outline-none dark:focus-visible:outline-none focus-visible:before:ring-inset focus-visible:before:ring-2',
        childLinkWrapper: 'min-w-0',
        childLinkIcon: 'size-5 shrink-0',
        childLinkLabel: 'truncate',
        childLinkLabelExternalIcon: 'inline-block size-3 align-top text-dimmed',
        childLinkDescription: 'text-muted',
        separator: 'px-2 h-px bg-border',
        viewportWrapper: 'absolute top-full left-0 flex w-full',
        viewport: 'relative overflow-hidden bg-default shadow-lg rounded-md ring ring-default h-(--reka-navigation-menu-viewport-height) w-full transition-[width,height,left] duration-200 origin-[top_center] data-[state=open]:animate-[scale-in_100ms_ease-out] data-[state=closed]:animate-[scale-out_100ms_ease-in] z-[1]',
        content: '',
        indicator: 'absolute data-[state=visible]:animate-[fade-in_100ms_ease-out] data-[state=hidden]:animate-[fade-out_100ms_ease-in] data-[state=hidden]:opacity-0 bottom-0 z-[2] w-(--reka-navigation-menu-indicator-size) translate-x-(--reka-navigation-menu-indicator-position) flex h-2.5 items-end justify-center overflow-hidden transition-[translate,width] duration-200',
        arrow: 'relative top-[50%] size-2.5 rotate-45 border border-default bg-default z-[1] rounded-xs'
      },
      variants: {
        color: {
          primary: {
            link: 'focus-visible:before:ring-white',
            childLink: 'focus-visible:before:ring-white'
          },
          secondary: {
            link: 'focus-visible:before:ring-secondary',
            childLink: 'focus-visible:before:ring-secondary'
          },
          success: {
            link: 'focus-visible:before:ring-success',
            childLink: 'focus-visible:before:ring-success'
          },
          info: {
            link: 'focus-visible:before:ring-info',
            childLink: 'focus-visible:before:ring-info'
          },
          warning: {
            link: 'focus-visible:before:ring-warning',
            childLink: 'focus-visible:before:ring-warning'
          },
          error: {
            link: 'focus-visible:before:ring-error',
            childLink: 'focus-visible:before:ring-error'
          },
          neutral: {
            link: 'focus-visible:before:ring-inverted',
            childLink: 'focus-visible:before:ring-inverted'
          }
        },
        highlightColor: {
          primary: '',
          secondary: '',
          success: '',
          info: '',
          warning: '',
          error: '',
          neutral: ''
        },
        variant: {
          pill: '',
          link: ''
        },
        orientation: {
          horizontal: {
            root: 'items-center justify-between',
            list: 'flex items-center',
            item: 'py-2',
            link: 'px-2.5 py-1.5 before:inset-x-px before:inset-y-0',
            childList: 'grid p-2',
            childLink: 'px-3 py-2 gap-2 before:inset-x-px before:inset-y-0',
            childLinkLabel: 'font-medium',
            content: 'absolute top-0 left-0 w-full max-h-[70vh] overflow-y-auto'
          },
          vertical: {
            root: 'flex-col',
            link: 'flex-row px-2.5 py-1.5 before:inset-y-px before:inset-x-0',
            childLabel: 'px-1.5 py-0.5',
            childLink: 'p-1.5 gap-1.5 before:inset-y-px before:inset-x-0'
          }
        },
        contentOrientation: {
          horizontal: {
            viewportWrapper: 'justify-center',
            content: 'data-[motion=from-start]:animate-[enter-from-left_200ms_ease] data-[motion=from-end]:animate-[enter-from-right_200ms_ease] data-[motion=to-start]:animate-[exit-to-left_200ms_ease] data-[motion=to-end]:animate-[exit-to-right_200ms_ease]'
          },
          vertical: {
            viewport: 'sm:w-(--reka-navigation-menu-viewport-width) left-(--reka-navigation-menu-viewport-left)'
          }
        },
        active: {
          true: {
            childLink: 'before:bg-elevated',
            childLinkIcon: 'text-default'
          },
          false: {
            link: 'text-muted',
            linkLeadingIcon: 'text-dimmed',
            childLink: [
              'hover:before:bg-elevated/50 text-default hover:text-highlighted',
              'transition-colors before:transition-colors'
            ],
            childLinkIcon: [
              'text-dimmed group-hover:text-default',
              'transition-colors'
            ]
          }
        },
        disabled: {
          true: {
            link: 'cursor-not-allowed opacity-75'
          }
        },
        highlight: {
          true: ''
        },
        level: {
          true: ''
        },
        collapsed: {
          true: ''
        }
      },
      compoundVariants: [
        {
          orientation: 'horizontal',
          contentOrientation: 'horizontal',
          class: {
            childList: 'grid-cols-2 gap-2'
          }
        },
        {
          orientation: 'horizontal',
          contentOrientation: 'vertical',
          class: {
            childList: 'gap-1',
            content: 'w-60'
          }
        },
        {
          orientation: 'vertical',
          collapsed: false,
          class: {
            childList: 'ms-5 border-s border-default',
            childItem: 'ps-1.5 -ms-px',
            content: 'data-[state=open]:animate-[collapsible-down_200ms_ease-out] data-[state=closed]:animate-[collapsible-up_200ms_ease-out] overflow-hidden'
          }
        },
        {
          orientation: 'vertical',
          collapsed: true,
          class: {
            link: 'px-1.5',
            linkLabel: 'hidden',
            linkTrailing: 'hidden',
            content: 'shadow-sm rounded-sm min-h-6 p-1'
          }
        },
        {
          orientation: 'horizontal',
          highlight: true,
          class: {
            link: [
              'after:absolute after:-bottom-2 after:inset-x-2.5 after:block after:h-px after:rounded-full',
              'after:transition-colors'
            ]
          }
        },
        {
          orientation: 'vertical',
          highlight: true,
          level: true,
          class: {
            link: [
              'after:absolute after:-start-1.5 after:inset-y-0.5 after:block after:w-px after:rounded-full',
              'after:transition-colors'
            ]
          }
        },
        {
          disabled: false,
          active: false,
          variant: 'pill',
          class: {
            link: [
              'hover:text-highlighted hover:before:bg-elevated/50',
              'transition-colors before:transition-colors'
            ],
            linkLeadingIcon: [
              'group-hover:text-default',
              'transition-colors'
            ]
          }
        },
        {
          disabled: false,
          active: false,
          variant: 'pill',
          orientation: 'horizontal',
          class: {
            link: 'data-[state=open]:text-highlighted',
            linkLeadingIcon: 'group-data-[state=open]:text-default'
          }
        },
        {
          disabled: false,
          variant: 'pill',
          highlight: true,
          orientation: 'horizontal',
          class: {
            link: 'data-[state=open]:before:bg-elevated/50'
          }
        },
        {
          disabled: false,
          variant: 'pill',
          highlight: false,
          active: false,
          orientation: 'horizontal',
          class: {
            link: 'data-[state=open]:before:bg-elevated/50'
          }
        },
        {
          color: 'primary',
          variant: 'pill',
          active: true,
          class: {
            link: 'text-primary',
            linkLeadingIcon: 'text-primary group-data-[state=open]:text-primary'
          }
        },
        {
          color: 'neutral',
          variant: 'pill',
          active: true,
          class: {
            link: 'text-highlighted',
            linkLeadingIcon: 'text-highlighted group-data-[state=open]:text-highlighted'
          }
        },
        {
          variant: 'pill',
          active: true,
          highlight: false,
          class: {
            link: 'before:bg-elevated'
          }
        },
        {
          variant: 'pill',
          active: true,
          highlight: true,
          disabled: false,
          class: {
            linkLeadingIcon: 'group-data-[state=open]:text-default text-highlighted',
            link: [
              'hover:before:bg-elevated/50',
              "bg-primary",
              'text-highlighted',
              'before:transition-colors'
            ]
          }
        },
        {
          disabled: false,
          active: false,
          variant: 'link',
          class: {
            link: [
              'hover:text-highlighted',
              'transition-colors'
            ],
            linkLeadingIcon: [
              'group-hover:text-default',
              'transition-colors'
            ]
          }
        },
        {
          disabled: false,
          active: false,
          variant: 'link',
          orientation: 'horizontal',
          class: {
            link: 'data-[state=open]:text-highlighted',
            linkLeadingIcon: 'group-data-[state=open]:text-default'
          }
        },
        {
          color: 'primary',
          variant: 'link',
          active: true,
          class: {
            link: 'text-primary',
            linkLeadingIcon: 'text-primary group-data-[state=open]:text-primary'
          }
        },
        {
          color: 'neutral',
          variant: 'link',
          active: true,
          class: {
            link: 'text-highlighted',
            linkLeadingIcon: 'text-highlighted group-data-[state=open]:text-highlighted'
          }
        },
        {
          highlightColor: 'primary',
          highlight: true,
          level: true,
          active: true,
          class: {
            link: 'after:bg-primary'
          }
        },
        {
          highlightColor: 'neutral',
          highlight: true,
          level: true,
          active: true,
          class: {
            link: 'after:bg-inverted'
          }
        }
      ],
      defaultVariants: {
        color: 'primary',
        highlightColor: 'primary',
        variant: 'pill'
      }
    },
    progress: {
      variants: {
        animation: {
          carousel: "",
          "carousel-inverse": "",
          swing: "",
          elastic: "",
        },
        color: {
          primary: {
            indicator: "bg-primary",
            steps: "text-primary",
          },
          secondary: {
            indicator: "bg-secondary",
            steps: "text-secondary",
          },
          success: {
            indicator: "bg-success",
            steps: "text-success",
          },
          info: {
            indicator: "bg-info",
            steps: "text-info",
          },
          warning: {
            indicator: "bg-warning",
            steps: "text-warning",
          },
          error: {
            indicator: "bg-error",
            steps: "text-error",
          },
          neutral: {
            indicator: "bg-inverted",
            steps: "text-inverted",
          },
        },
        size: {
          "2xs": {
            status: "text-xs",
            steps: "text-xs",
          },
          xs: {
            status: "text-xs",
            steps: "text-xs",
          },
          sm: {
            status: "text-sm",
            steps: "text-sm",
          },
          md: {
            status: "text-sm",
            steps: "text-sm",
          },
          lg: {
            status: "text-sm",
            steps: "text-sm",
          },
          xl: {
            status: "text-base",
            steps: "text-base",
          },
          "2xl": {
            status: "text-base",
            steps: "text-base",
          },
        },
      },
    },
    switch: {
      slots: {
        root: 'relative flex items-start',
        base: [
          'inline-flex items-center shrink-0 rounded-full border-2 border-transparent focus-visible:outline-2 focus-visible:outline-offset-2 data-[state=unchecked]:bg-accented',
          'transition-[background] duration-200'
        ],
        container: 'flex items-center',
        thumb: 'group pointer-events-none rounded-full bg-default shadow-lg ring-0 transition-transform duration-200 data-[state=unchecked]:translate-x-0 data-[state=unchecked]:rtl:-translate-x-0 flex items-center justify-center',
        icon: [
          'absolute shrink-0 group-data-[state=unchecked]:text-dimmed opacity-0 size-10/12',
          'transition-[color,opacity] duration-200'
        ],
        wrapper: 'ms-2',
        label: 'block font-medium text-default',
        description: 'text-muted'
      },
      variants: {
        color: {
          primary: {
            base: 'data-[state=checked]:bg-primary focus-visible:outline-primary',
            icon: 'group-data-[state=checked]:text-primary'
          },
          secondary: {
            base: 'data-[state=checked]:bg-secondary focus-visible:outline-secondary',
            icon: 'group-data-[state=checked]:text-secondary'
          },
          success: {
            base: 'data-[state=checked]:bg-success focus-visible:outline-success',
            icon: 'group-data-[state=checked]:text-success'
          },
          info: {
            base: 'data-[state=checked]:bg-info focus-visible:outline-info',
            icon: 'group-data-[state=checked]:text-info'
          },
          warning: {
            base: 'data-[state=checked]:bg-warning focus-visible:outline-warning',
            icon: 'group-data-[state=checked]:text-warning'
          },
          error: {
            base: 'data-[state=checked]:bg-error focus-visible:outline-error',
            icon: 'group-data-[state=checked]:text-error'
          },
          neutral: {
            base: 'data-[state=checked]:bg-inverted focus-visible:outline-inverted',
            icon: 'group-data-[state=checked]:text-highlighted'
          }
        },
        size: {
          xs: {
            base: 'w-7',
            container: 'h-4',
            thumb: 'size-3 data-[state=checked]:translate-x-3 data-[state=checked]:rtl:-translate-x-3',
            wrapper: 'text-xs'
          },
          sm: {
            base: 'w-8',
            container: 'h-4',
            thumb: 'size-3.5 data-[state=checked]:translate-x-3.5 data-[state=checked]:rtl:-translate-x-3.5',
            wrapper: 'text-xs'
          },
          md: {
            base: 'w-9',
            container: 'h-5',
            thumb: 'size-4 data-[state=checked]:translate-x-4 data-[state=checked]:rtl:-translate-x-4',
            wrapper: 'text-sm'
          },
          lg: {
            base: 'w-10',
            container: 'h-5',
            thumb: 'size-4.5 data-[state=checked]:translate-x-4.5 data-[state=checked]:rtl:-translate-x-4.5',
            wrapper: 'text-sm'
          },
          xl: {
            base: 'w-11',
            container: 'h-6',
            thumb: 'size-5 data-[state=checked]:translate-x-5 data-[state=checked]:rtl:-translate-x-5',
            wrapper: 'text-base'
          }
        },
        checked: {
          true: {
            icon: 'group-data-[state=checked]:opacity-100'
          }
        },
        unchecked: {
          true: {
            icon: 'group-data-[state=unchecked]:opacity-100'
          }
        },
        loading: {
          true: {
            icon: 'animate-spin'
          }
        },
        required: {
          true: {
            label: "after:content-['*'] after:ms-0.5 after:text-error"
          }
        },
        disabled: {
          true: {
            root: 'opacity-75',
            base: 'cursor-not-allowed',
            label: 'cursor-not-allowed',
            description: 'cursor-not-allowed'
          }
        }
      },
      defaultVariants: {
        color: 'primary',
        size: 'md'
      }
    },

    button: {
      block: {
        true: {
          base: "w-full justify-center",
          trailingIcon: "ms-auto",
        },
      },
      slots: {
        leadingIcon: 'shrink-0 font-semibold',
      },
      size: {
        xs: {
          base: "px-2 py-1 text-xs gap-1",
          leadingIcon: "size-4",
          leadingAvatarSize: "3xs",
          trailingIcon: "size-4",
        },
        sm: {
          base: "px-2.5 py-1.5 text-xs gap-1.5",
          leadingIcon: "size-4",
          leadingAvatarSize: "3xs",
          trailingIcon: "size-4",
        },
        md: {
          base: "px-2.5 py-1.5 text-sm gap-1.5",
          leadingIcon: "size-5",
          leadingAvatarSize: "2xs",
          trailingIcon: "size-5",
        },
        lg: {
          base: "px-3 py-2 text-sm gap-2",
          leadingIcon: "size-5",
          leadingAvatarSize: "2xs",
          trailingIcon: "size-5",
        },
        xl: {
          base: "px-3 py-2 text-base gap-2",
          leadingIcon: "size-6",
          leadingAvatarSize: "xs",
          trailingIcon: "size-6",
        },
      },
      compoundVariants: [
        {
          color: "primary",
          variant: "solid",
          class:
            "text-inverted font-semibold tracking-wider from-primary-300 to-primary-400 bg-gradient-to-br hover:bg-primary/75 active:bg-primary/75 disabled:bg-primary aria-disabled:bg-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary text-secondary-900 shadow-sm",
        },
        {
          color: "primary",
          variant: "outline",
          class:
            "ring ring-inset ring-primary/50 text-primary hover:bg-primary/10 active:bg-primary/10 disabled:bg-transparent aria-disabled:bg-transparent dark:disabled:bg-transparent dark:aria-disabled:bg-transparent focus:outline-none focus-visible:ring-2 focus-visible:ring-primary",
        },
        {
          color: "primary",
          variant: "soft",
          class:
            "text-primary bg-primary/10 hover:bg-primary/15 active:bg-primary/15 focus:outline-none focus-visible:bg-primary/15 disabled:bg-primary/10 aria-disabled:bg-primary/10",
        },
        {
          color: "primary",
          variant: "subtle",
          class:
            "text-primary ring ring-inset ring-primary/25 bg-primary/10 hover:bg-primary/15 active:bg-primary/15 disabled:bg-primary/10 aria-disabled:bg-primary/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary",
        },
        {
          color: "primary",
          variant: "ghost",
          class:
            "text-primary hover:bg-primary/10 active:bg-primary/10 focus:outline-none focus-visible:bg-primary/10 disabled:bg-transparent aria-disabled:bg-transparent dark:disabled:bg-transparent dark:aria-disabled:bg-transparent",
        },
        {
          color: "primary",
          variant: "link",
          class:
            "text-primary hover:text-primary/75 active:text-primary/75 disabled:text-primary aria-disabled:text-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary",
        },
        {
          color: "error",
          variant: "ghost",
          class:
            "text-error-500 bg-transparent font-semibold",
        },
        {
          color: "neutral",
          variant: "solid",
          class:
            "text-inverted bg-inverted hover:bg-inverted/90 active:bg-inverted/90 disabled:bg-inverted aria-disabled:bg-inverted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-inverted",
        },
        {
          color: "neutral",
          variant: "outline",
          class:
            "ring ring-inset ring-accented text-default bg-default hover:bg-elevated active:bg-elevated disabled:bg-default aria-disabled:bg-default focus:outline-none focus-visible:ring-2 focus-visible:ring-inverted",
        },
        {
          color: "neutral",
          variant: "soft",
          class:
            "text-default bg-elevated hover:bg-accented/75 active:bg-accented/75 focus:outline-none focus-visible:bg-accented/75 disabled:bg-elevated aria-disabled:bg-elevated",
        },
        {
          color: "neutral",
          variant: "subtle",
          class:
            "ring ring-inset ring-accented text-default bg-elevated hover:bg-accented/75 active:bg-accented/75 disabled:bg-elevated aria-disabled:bg-elevated focus:outline-none focus-visible:ring-2 focus-visible:ring-inverted",
        },
        {
          color: "neutral",
          variant: "ghost",
          class:
            "text-default hover:bg-elevated active:bg-elevated focus:outline-none focus-visible:bg-elevated hover:disabled:bg-transparent dark:hover:disabled:bg-transparent hover:aria-disabled:bg-transparent dark:hover:aria-disabled:bg-transparent",
        },
        {
          color: "neutral",
          variant: "link",
          class:
            "text-muted hover:text-default active:text-default disabled:text-muted aria-disabled:text-muted focus:outline-none focus-visible:ring-inset focus-visible:ring-2 focus-visible:ring-inverted",
        },
        {
          size: "xs",
          square: true,
          class: "p-1",
        },
        {
          size: "sm",
          square: true,
          class: "p-1.5",
        },
        {
          size: "md",
          square: true,
          class: "p-1.5",
        },
        {
          size: "lg",
          square: true,
          class: "p-2",
        },
        {
          size: "xl",
          square: true,
          class: "p-2",
        },
        {
          loading: true,
          leading: true,
          class: {
            leadingIcon: "animate-spin",
          },
        },
        {
          loading: true,
          leading: false,
          trailing: true,
          class: {
            trailingIcon: "animate-spin",
          },
        },
      ],
      variants: {
        solid: {
          bg: "bg-primary-500",
        },
      },
    },
  },
});
