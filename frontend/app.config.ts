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
      check: "heroicons:check",
      loading: "heroicons:arrow-path",
      charter: "heroicons:clipboard-document-list",
      warning: "heroicons:exclamation-triangle",
      alert: "heroicons:alert-circle",
      calendar: "heroicons:calendar",
      download: "heroicons:arrow-down-tray",
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
    },
  },
  theme: {
    dark: false,
    primaryColor: "primary",
  },

  ui: {
    badge: {
      slots: {
        base: "font-medium inline-flex items-center",
        label: "truncate",
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
        label: "block text-black",
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
      base: "max-w-(--ui-container) w-full space-y-8 bg-neutral-200 px-8 py-4 rounded-xl z-100 block relative shadow-xl",
    },
    card: {
      slots: {
        root: "rounded-lg overflow-hidden",
        header: "p-4 sm:px-6 border-none",
        body: "p-4 sm:p-6 border-none",
        footer: "p-4 sm:px-6 border-none",
      },
      variants: {
        variant: {
          solid: {
            root: "bg-inverted text-inverted",
          },
          outline: {
            root: "from-secondary-50 to-secondary-100 bg-gradient-to-r ring-2 from ring ring-primary-500 divide-y divide-default",
          },
          soft: {
            root: "bg-secondary-50 rounded-xl z-100 block relative shadow-xl",
          },
          subtle: {
            root: "bg-elevated/50 ring ring-default divide-y divide-default",
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
        wrapper: "",
        labelWrapper:
          "flex content-start items-start justify-start min-w-[50%]",
        label: "block font-medium text-default text-black min-w-max",
        container: "relative ",
        description: "text-muted",
        error: "mt-2 text-error",
        hint: "text-muted",
        help: "mt-2 text-muted",
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
            root: "flex flex-row justify-between place-items-baseline gap-32",
          },
        },
      },
      defaultVariants: {
        size: "md",
        orientation: "vertical",
      },
    },
    main: {
      base: "min-h-[calc(100vh-var(--ui-header-height))] max-w-8xl px-24 py-16",
    },
    modal: {
      slots: {
        overlay: "fixed inset-0",
        content:
          "bg-secondary-50 divide-y divide-default flex flex-col focus:outline-none",
        header: "flex items-center gap-1.5 p-4 sm:px-6 min-h-16",
        wrapper: "",
        body: "flex-1 p-4 sm:p-6",
        footer: "flex flex-row justify-between gap-1.5 p-4 sm:px-6",
        title: "text-highlighted font-semibold text-2xl text-black",
        description: "mt-1 text-muted text-sm",
        close: "absolute top-4 end-4",
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
