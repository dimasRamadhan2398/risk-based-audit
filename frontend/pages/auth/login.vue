<template>
  <div
    class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8"
  >
    <div
      class="absolute bg-[url('../../../assets/images/audits/bg-login.jpg')] bg-cover bg-center bg-transparent w-full h-full block opacity-10 z-20"
    />
    <div class="absolute w-full h-full bg-opacity-30 z-10" />
    <UContainer
      class="max-w-md w-full space-y-8 bg-neutral-400 px-8 py-4 rounded-xl z-40 block relative shadow-2xl"
    >
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          {{ t("auth.login.title") }}
        </h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          {{ t("auth.login.subtitle") }}
        </p>
      </div>
      <UForm
        ref="form"
        :schema="schema"
        :state="state"
        class="mt-8 space-y-6"
        @submit="handleLogin"
      >
        <UFormField :label="t('auth.email')" name="email">
          <UInput
            id="email-address"
            class="w-full"
            name="email"
            type="email"
            v-model="state.email"
            autocomplete="email"
            color="neutral"
            highlight
            variant="ghost"
            :placeholder="t('auth.email')"
          >
          </UInput>
        </UFormField>
        <UFormField :label="t('auth.password')" name="password">
          <UInput
            id="password"
            name="password"
            type="password"
            v-model="state.password"
            color="neutral"
            highlight
            class="w-full"
            variant="ghost"
            :placeholder="t('auth.password')"
          >
          </UInput>
        </UFormField>

        <div v-if="error" class="rounded-md bg-red-50 p-4">
          <p class="text-sm text-red-800">{{ error }}</p>
        </div>

        <div>
          <UButton
            :label="loading ? t('auth.loggingIn') : t('auth.login.button')"
            type="submit"
            :block="true"
            color="secondary"
            :loading="loading"
          >
          </UButton>
        </div>

        <div class="text-center">
          <p class="text-sm text-gray-600">
            {{ t("auth.noAccount") }} {{ t("auth.contactAdmin") }}
          </p>
        </div>
      </UForm>
    </UContainer>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "~/stores/auth";
import { useI18n } from "~/composables/useI18n";
import * as z from "zod";
import type { FormSubmitEvent } from "@nuxt/ui";

definePageMeta({
  layout: "auth",
  pageTransition: {
    name: "fade",
    mode: "in-out",
    type: "animation",
    duration: 500,
    appear: true,
  },
  viewTransition: true,
  scrollToTop: true,
  middleware: "guest",
});

const { t } = useI18n();
const authStore = useAuthStore();
const router = useRouter();

type Schema = z.output<typeof schema>;

const schema = z.object({
  email: z
    .string()
    .min(1, "Please enter your email")
    .email("Please enter a valid email"),
  password: z.string().min(1, "Please enter your password"),
});

const state = reactive<Partial<Schema>>({
  email: "",
  password: "",
});

const loading = ref(false);
const error = ref("");

const handleLogin = async (event: FormSubmitEvent<Schema>) => {
  loading.value = true;
  error.value = "";

  try {
    await authStore.loginDummy(event.data);
    router.push("/dashboard");
  } catch (err: any) {
    error.value = err.message || t("auth.login.error");
  } finally {
    loading.value = false;
    console.log(event.data);
  }
};
</script>
