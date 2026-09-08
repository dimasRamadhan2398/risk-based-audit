<template>
  <div class="min-h-screen bg-[var(--bg-main)] text-[var(--text-main)] transition-colors duration-300 flex flex-col">
    <!-- Standalone Header Navigation -->
    <header class="sticky top-0 z-40 w-full border-b border-[var(--border-main)] bg-[var(--bg-surface)]/85 backdrop-blur-md">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
        <!-- Brand & Context -->
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <Logo class="h-7 w-auto" />
          </div>

          <div class="h-5 w-px bg-[var(--border-main)] hidden sm:block" />

          <div class="hidden sm:flex items-center gap-2">
            <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              <UIcon
                name="i-lucide-globe"
                class="w-3.5 h-3.5"
              />
              Client Site Generator
            </span>
            <span class="inline-flex items-center gap-1.5 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
              <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              Gateway Orchestrator Online
            </span>
          </div>
        </div>

        <!-- Right Header Actions (Without Back to Dashboard) -->
        <div class="flex items-center gap-3">
          <span class="text-xs text-[var(--text-muted)] font-mono hidden md:inline">
            Domain Root: *.auditsphere.id
          </span>
          <UColorModeButton />
        </div>
      </div>
    </header>

    <!-- Main Content Area: Centered Vertically -->
    <main class="flex-1 flex items-center justify-center p-4 sm:p-6 lg:p-8">
      <!-- Generate New Client Site Card Box -->
      <div class="w-full max-w-4xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden flex flex-col transition-all">
        <!-- SUCCESS STATE -->
        <div
          v-if="isDeploymentComplete && deployedSite"
          class="p-8 sm:p-10 space-y-6 text-center"
        >
          <div class="w-16 h-16 mx-auto rounded-2xl bg-emerald-50 dark:bg-emerald-950/60 border border-emerald-200 dark:border-emerald-800 flex items-center justify-center text-emerald-600 dark:text-emerald-400 shadow-sm animate-bounce">
            <UIcon
              name="i-lucide-check-circle"
              class="w-10 h-10"
            />
          </div>

          <div class="space-y-2 max-w-lg mx-auto">
            <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
              Client Site Successfully Generated!
            </h2>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              The isolated instance for <span class="font-bold text-gray-900 dark:text-white">{{ deployedSite.clientName }}</span> has been provisioned and is ready for onboarding.
            </p>
          </div>

          <!-- Provisioned Details Card -->
          <div class="max-w-xl mx-auto p-5 rounded-2xl bg-gray-50 dark:bg-gray-850 border border-gray-200 dark:border-gray-800 text-left space-y-3.5 text-xs">
            <div class="flex items-center justify-between border-b pb-2.5 border-gray-200 dark:border-gray-800">
              <span class="text-gray-500 dark:text-gray-400 font-medium">Portal URL</span>
              <a
                :href="`https://${deployedSite.domain}`"
                target="_blank"
                class="font-mono font-bold text-primary hover:underline flex items-center gap-1 text-sm"
              >
                https://{{ deployedSite.domain }}
                <UIcon
                  name="i-lucide-external-link"
                  class="w-3.5 h-3.5"
                />
              </a>
            </div>

            <div class="flex items-center justify-between border-b pb-2.5 border-gray-200 dark:border-gray-800">
              <span class="text-gray-500 dark:text-gray-400 font-medium">API Gateway</span>
              <span class="font-mono text-gray-700 dark:text-gray-300">https://{{ deployedSite.apiDomain }}</span>
            </div>

            <div class="flex items-center justify-between border-b pb-2.5 border-gray-200 dark:border-gray-800">
              <span class="text-gray-500 dark:text-gray-400 font-medium">Container Ports</span>
              <span class="font-mono text-gray-700 dark:text-gray-300">Frontend: {{ deployedSite.frontendPort }} | Kong: {{ deployedSite.kongPort }}</span>
            </div>

            <div class="flex items-center justify-between border-b pb-2.5 border-gray-200 dark:border-gray-800">
              <span class="text-gray-500 dark:text-gray-400 font-medium">Chief Audit Executive</span>
              <span class="text-gray-800 dark:text-gray-200 font-medium">{{ deployedSite.adminName }} ({{ deployedSite.adminEmail }})</span>
            </div>

            <div class="flex items-center justify-between">
              <span class="text-gray-500 dark:text-gray-400 font-medium">Compliance Framework</span>
              <span class="text-gray-800 dark:text-gray-200 font-medium">{{ deployedSite.complianceFramework }}</span>
            </div>
          </div>

          <!-- Automated CLI Command -->
          <div class="max-w-xl mx-auto text-left space-y-1.5">
            <div class="flex items-center justify-between text-xs">
              <span class="text-gray-500 font-medium">CLI Deployment Command</span>
              <button
                type="button"
                class="text-primary hover:text-primary-700 font-medium flex items-center gap-1"
                @click="copyText(generatedCliCommand, 'CLI Command')"
              >
                <UIcon
                  name="i-lucide-copy"
                  class="w-3.5 h-3.5"
                />
                Copy
              </button>
            </div>
            <div class="bg-gray-950 text-emerald-400 p-3.5 rounded-xl font-mono text-xs overflow-x-auto border border-gray-800">
              <code>{{ generatedCliCommand }}</code>
            </div>
          </div>

          <div class="pt-2 flex justify-center gap-3">
            <ReusableButton
              variant="fill"
              color="primary"
              size="md"
              icon="i-lucide-plus"
              @click="resetForm"
            >
              Generate Another Site
            </ReusableButton>
          </div>
        </div>

        <!-- WIZARD STEP FORM -->
        <template v-else>
          <!-- Card Header -->
          <div class="px-6 sm:px-8 py-5 dark:bg-primary-600 dark:text-secondary flex items-center justify-between">
            <div class="flex items-center gap-3.5">
              <div class="w-10 h-10 rounded-xl bg-primary-50 dark:bg-primary-950/60 border border-primary-200 dark:border-primary-800 flex items-center justify-center text-primary-600 dark:text-primary-400 shrink-0 shadow-sm">
                <UIcon
                  name="i-lucide-globe"
                  class="w-5 h-5"
                />
              </div>
              <div>
                <h2 class="text-lg sm:text-xl font-bold text-gray-900 dark:text-neutral-200!">
                  Generate New Client Site
                </h2>
                <p class="text-xs text-gray-500 dark:text-neutral-200!">
                  Provision an isolated AuditSphere instance
                </p>
              </div>
            </div>

            <span class="text-xs font-semibold px-2.5 py-1 rounded-full bg-primary/50 text-secondary-700 border border-primary/20">
              Step {{ currentStep }} of {{ steps.length }}
            </span>
          </div>

          <!-- Step Navigation Stepper -->
          <div class="px-6 sm:px-8 py-3 bg-gray-50/70 dark:bg-gray-800/40 border-b border-gray-100 dark:border-gray-800">
            <div class="flex items-center justify-between">
              <div
                v-for="(step, idx) in steps"
                :key="step.id"
                class="flex items-center gap-2 cursor-pointer select-none"
                @click="currentStep > idx ? currentStep = idx + 1 : null"
              >
                <div
                  :class="[
                    'w-6 h-6 rounded-full text-xs font-bold flex items-center justify-center transition-all duration-200',
                    currentStep === idx + 1
                      ? 'bg-primary text-white shadow-sm ring-2 ring-primary/20'
                      : currentStep > idx + 1
                        ? 'bg-emerald-500 text-white'
                        : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400'
                  ]"
                >
                  <UIcon
                    v-if="currentStep > idx + 1"
                    name="i-lucide-check"
                    class="w-3.5 h-3.5"
                  />
                  <span v-else>{{ idx + 1 }}</span>
                </div>
                <span
                  :class="[
                    'text-xs font-medium hidden sm:inline transition-colors',
                    currentStep === idx + 1 ? 'text-gray-900 dark:text-white font-semibold' : 'text-gray-500 dark:text-gray-400'
                  ]"
                >
                  {{ step.title }}
                </span>
                <UIcon
                  v-if="idx < steps.length - 1"
                  name="i-lucide-chevron-right"
                  class="w-3.5 h-3.5 text-gray-300 dark:text-gray-600 hidden sm:inline ml-2"
                />
              </div>
            </div>
          </div>

          <!-- Card Body -->
          <div class="p-6 sm:p-8 space-y-6 flex-1 overflow-y-auto max-h-[62vh]">
            <!-- STEP 1: Organization & Subdomain -->
            <div
              v-if="currentStep === 1"
              class="space-y-5"
            >
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <AppFormField
                  label="Client Organization Name"
                  tooltip="Official name of the client organization displayed across their portal"
                  required
                  counter
                  :max-count="100"
                  :model-value="form.clientName"
                >
                  <UInput
                    v-model="form.clientName"
                    placeholder="Company Name"
                    maxlength="100"
                    class="w-full"
                  />
                </AppFormField>

                <AppFormField
                  label="Industry / Sector"
                  tooltip="Primary industry sector for benchmarking risk metrics"
                  required
                >
                  <ReusableSelectMenu
                    v-model="form.industry"
                    :items="industryOptions"
                    placeholder="Select industry sector"
                    class="w-full"
                  />
                </AppFormField>
              </div>

              <!-- Subdomain Slug & Live URL Preview -->
              <div class="space-y-2">
                <AppFormField
                  label="Tenant Subdomain Slug"
                  tooltip="Alphanumeric identifier used for DNS subdomain, database prefixes, and container routing"
                  required
                  counter
                  :max-count="40"
                  :model-value="form.slug"
                >
                  <div class="flex items-center">
                    <span class="inline-flex items-center px-3 py-2 rounded-l-md border border-r-0 border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 text-sm font-mono select-none">
                      https://
                    </span>
                    <UInput
                      v-model="form.slug"
                      placeholder="accenture"
                      maxlength="40"
                      class="flex-1 rounded-none"
                      :ui="{ base: 'rounded-none font-mono text-primary font-semibold' }"
                      @input="sanitizeSlug"
                    />
                    <span class="inline-flex items-center px-3 py-2 rounded-r-md border border-l-0 border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-600 dark:text-gray-300 text-sm font-mono font-medium select-none">
                      .auditsphere.id
                    </span>
                  </div>
                </AppFormField>

                <!-- Live Target Domain Cards Preview -->
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mt-3">
                  <div class="p-3 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-850/70 flex items-start gap-2.5">
                    <UIcon
                      name="i-lucide-laptop"
                      class="w-4 h-4 text-primary shrink-0 mt-0.5"
                    />
                    <div class="text-xs">
                      <div class="text-gray-400 font-medium">
                        Frontend Portal Domain
                      </div>
                      <div class="font-mono font-semibold text-gray-900 dark:text-white mt-0.5 break-all">
                        {{ targetDomain }}
                      </div>
                    </div>
                  </div>

                  <div class="p-3 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-850/70 flex items-start gap-2.5">
                    <UIcon
                      name="i-lucide-network"
                      class="w-4 h-4 text-secondary shrink-0 mt-0.5"
                    />
                    <div class="text-xs">
                      <div class="text-gray-400 font-medium">
                        Kong API Gateway Endpoint
                      </div>
                      <div class="font-mono font-semibold text-gray-900 dark:text-white mt-0.5 break-all">
                        {{ targetApiDomain }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Brand Customization & Framework -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2">
                <AppFormField
                  label="Primary Brand Color"
                  tooltip="Used for accent highlights on the client's custom theme"
                  optional
                >
                  <div class="flex items-center gap-3">
                    <input
                      v-model="form.brandColor"
                      type="color"
                      class="w-10 h-10 rounded-lg cursor-pointer border border-gray-300 dark:border-gray-700 p-0.5 bg-white dark:bg-gray-800"
                    >
                    <UInput
                      v-model="form.brandColor"
                      placeholder="#0284c7"
                      class="flex-1 font-mono uppercase"
                      maxlength="7"
                    />
                  </div>
                </AppFormField>

                <AppFormField
                  label="Compliance Framework"
                  tooltip="Pre-seeds audit charter guidelines and risk appetite templates"
                  required
                >
                  <ReusableSelectMenu
                    v-model="form.complianceFramework"
                    :items="frameworkOptions"
                    placeholder="Select framework"
                    class="w-full"
                  />
                </AppFormField>
              </div>
            </div>

            <!-- STEP 2: Administrator & Security -->
            <div
              v-else-if="currentStep === 2"
              class="space-y-5"
            >
              <div class="p-3.5 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/40 flex items-center gap-3">
                <UIcon
                  name="i-lucide-shield-alert"
                  class="w-5 h-5 text-blue-600 dark:text-blue-400 shrink-0"
                />
                <p class="text-xs text-blue-800 dark:text-blue-300 leading-relaxed">
                  The initial administrator account will have <strong>Chief Audit Executive (CAE) / Super Admin</strong> privileges on the new tenant instance with automated RBAC seeding.
                </p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <AppFormField
                  label="Admin Full Name"
                  tooltip="Full name of the Chief Audit Executive or primary system administrator"
                  required
                  counter
                  :max-count="100"
                  :model-value="form.adminName"
                >
                  <UInput
                    v-model="form.adminName"
                    placeholder="e.g. Budi Santoso, CIA, CISA"
                    maxlength="100"
                    class="w-full"
                  />
                </AppFormField>

                <AppFormField
                  label="Admin Email Address"
                  tooltip="Will receive initial login credentials and security alerts"
                  required
                >
                  <UInput
                    v-model="form.adminEmail"
                    type="email"
                    placeholder="cae@example.com"
                    class="w-full"
                  />
                </AppFormField>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <AppFormField
                  label="Admin Phone / WhatsApp"
                  tooltip="Emergency contact for security and incident alerts"
                  optional
                >
                  <UInput
                    v-model="form.adminPhone"
                    placeholder="+62 812-3456-7890"
                    class="w-full"
                  />
                </AppFormField>

                <AppFormField
                  label="Two-Factor Authentication (2FA)"
                  tooltip="Enforce mandatory TOTP Two-Factor Authentication for all tenant auditors"
                >
                  <div class="pt-2">
                    <UCheckbox
                      v-model="form.enforceMfa"
                      label="Enforce Mandatory 2FA for all auditors"
                    />
                  </div>
                </AppFormField>
              </div>
            </div>

            <!-- STEP 3: Storage & Database Isolation -->
            <div
              v-else-if="currentStep === 3"
              class="space-y-5"
            >
              <!-- Auto-configuration Toggle Callout -->
              <div class="p-4 rounded-xl border border-primary-200 dark:border-primary-800/60 bg-gradient-to-r from-primary-50/70 to-blue-50/50 dark:from-primary-950/30 dark:to-blue-950/20 transition-all">
                <div class="flex items-start gap-3">
                  <UCheckbox
                    v-model="form.autoConfigInfrastructure"
                    class="mt-0.5"
                  />
                  <div class="space-y-1">
                    <label
                      class="text-xs font-bold text-gray-900 dark:text-white cursor-pointer select-none flex items-center gap-1.5"
                      @click="form.autoConfigInfrastructure = !form.autoConfigInfrastructure"
                    >
                      <UIcon
                        name="i-lucide-sparkles"
                        class="w-4 h-4 text-primary"
                      />
                      Let system automatically configure cloud infrastructure (Recommended)
                    </label>
                    <p class="text-xs text-gray-500 dark:text-gray-400 leading-relaxed">
                      Recommended for most client companies. The platform will automatically allocate container ports, configure the Google Drive evidence vault, assign host gateways, and prepare database clusters without requiring technical network setup.
                    </p>
                  </div>
                </div>
              </div>

              <!-- Automated Infrastructure Matrix View -->
              <div
                v-if="form.autoConfigInfrastructure"
                class="rounded-xl border border-dashed border-gray-200 dark:border-gray-800 bg-gray-50/70 dark:bg-gray-850/40 p-4 space-y-3"
              >
                <div class="flex items-center justify-between text-xs pb-2 border-b border-gray-200/60 dark:border-gray-800">
                  <span class="font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                    <UIcon
                      name="i-lucide-cpu"
                      class="w-4 h-4 text-emerald-500"
                    />
                    Automated Provisioning Matrix
                  </span>
                  <span class="px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800/40 flex items-center gap-1">
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                    Auto-Configured
                  </span>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                  <div class="space-y-1">
                    <span class="text-gray-400 block">Assigned Web Port</span>
                    <span class="font-mono font-medium text-gray-800 dark:text-gray-200">Port {{ form.frontendPort }} (Internal Nuxt)</span>
                  </div>
                  <div class="space-y-1">
                    <span class="text-gray-400 block">Assigned Gateway Port</span>
                    <span class="font-mono font-medium text-gray-800 dark:text-gray-200">Port {{ form.kongPort }} (Kong Gateway)</span>
                  </div>
                  <div class="space-y-1">
                    <span class="text-gray-400 block">Cloud Evidence Vault</span>
                    <span class="font-mono font-medium text-gray-800 dark:text-gray-200 truncate block">Dedicated Silo Vault (Auto-Allocated)</span>
                  </div>
                  <div class="space-y-1">
                    <span class="text-gray-400 block">Host Node Cluster</span>
                    <span class="font-mono font-medium text-gray-800 dark:text-gray-200">{{ form.serverIp }} (Production VPS)</span>
                  </div>
                </div>

                <div class="pt-1 text-[11px] text-gray-400 flex items-center gap-1">
                  <UIcon
                    name="i-lucide-info"
                    class="w-3.5 h-3.5 text-primary shrink-0"
                  />
                  <span>Uncheck the box above if your DevOps team requires custom container ports or specific Google Drive folder IDs.</span>
                </div>
              </div>

              <!-- Manual Configuration Inputs (Only when autoConfigInfrastructure is false) -->
              <template v-else>
                <AppFormField
                  label="Google Drive Storage Folder ID"
                  tooltip="Dedicated folder ID on Google Drive for storing interview recordings, working paper attachments, and observation evidence"
                  hint="From Google Drive URL"
                  required
                >
                  <UInput
                    v-model="form.gdriveFolderId"
                    placeholder="e.g. 1bX7yZ9kL0mN8pQ2rS4tU6vW8xYz1234A"
                    class="w-full font-mono text-xs"
                  />
                </AppFormField>

                <!-- Assigned Ports & Server IP -->
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <AppFormField
                    label="Host VPS IP"
                    tooltip="Public IP of the target hosting server running Docker and Nginx"
                    required
                  >
                    <UInput
                      v-model="form.serverIp"
                      placeholder="202.10.34.166"
                      class="w-full font-mono text-xs"
                    />
                  </AppFormField>

                  <AppFormField
                    label="Assigned Frontend Port"
                    tooltip="Internal container port for this tenant's Nuxt frontend"
                    required
                  >
                    <UInput
                      v-model.number="form.frontendPort"
                      type="number"
                      placeholder="3010"
                      class="w-full font-mono text-xs"
                    />
                  </AppFormField>

                  <AppFormField
                    label="Assigned Kong Proxy Port"
                    tooltip="Internal container port for this tenant's Kong Gateway"
                    required
                  >
                    <UInput
                      v-model.number="form.kongPort"
                      type="number"
                      placeholder="8090"
                      class="w-full font-mono text-xs"
                    />
                  </AppFormField>
                </div>
              </template>

              <!-- Isolated PostgreSQL Database Previews -->
              <div class="space-y-2">
                <label class="text-xs font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                  <UIcon
                    name="i-lucide-database"
                    class="w-3.5 h-3.5 text-primary"
                  />
                  Dedicated Isolated PostgreSQL Databases (Auto-Generated)
                </label>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs font-mono">
                  <div
                    v-for="dbName in generatedDatabases"
                    :key="dbName"
                    class="px-3 py-2 rounded-lg bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700/60 flex items-center justify-between"
                  >
                    <span>{{ dbName }}</span>
                    <UIcon
                      name="i-lucide-lock"
                      class="w-3.5 h-3.5 text-gray-400"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- STEP 4: Review & Deploy -->
            <div
              v-else-if="currentStep === 4"
              class="space-y-5"
            >
              <!-- Overview summary -->
              <div class="rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden text-xs">
                <div class="bg-gray-50 dark:bg-gray-800/80 px-4 py-2.5 font-semibold text-gray-900 dark:text-white border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
                  <span>Tenant Site Provisioning Summary</span>
                  <span class="text-emerald-600 dark:text-emerald-400 font-medium flex items-center gap-1">
                    <UIcon
                      name="i-lucide-check-circle-2"
                      class="w-3.5 h-3.5"
                    />
                    Validated
                  </span>
                </div>
                <div class="p-4 grid grid-cols-1 sm:grid-cols-2 gap-y-3 gap-x-6">
                  <div>
                    <span class="text-gray-400 block">Organization Name</span>
                    <span class="font-semibold text-gray-900 dark:text-white">{{ form.clientName }}</span>
                  </div>
                  <div>
                    <span class="text-gray-400 block">Target URL</span>
                    <span class="font-mono text-primary font-bold">https://{{ targetDomain }}</span>
                  </div>
                  <div>
                    <span class="text-gray-400 block">API Gateway</span>
                    <span class="font-mono text-gray-700 dark:text-gray-300">https://{{ targetApiDomain }}</span>
                  </div>
                  <div>
                    <span class="text-gray-400 block">Assigned Ports</span>
                    <span class="font-mono text-gray-700 dark:text-gray-300">FE: {{ form.frontendPort }} | Kong: {{ form.kongPort }}</span>
                  </div>
                  <div>
                    <span class="text-gray-400 block">Chief Audit Executive</span>
                    <span class="text-gray-700 dark:text-gray-300">{{ form.adminName }} ({{ form.adminEmail }})</span>
                  </div>
                  <div>
                    <span class="text-gray-400 block">Google Drive Storage</span>
                    <span class="font-mono text-gray-700 dark:text-gray-300 truncate block">{{ form.gdriveFolderId || 'Pending ID' }}</span>
                  </div>
                </div>
              </div>

              <!-- Automated CLI Command Box -->
              <div class="space-y-2">
                <div class="flex items-center justify-between">
                  <label class="text-xs font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                    <UIcon
                      name="i-lucide-terminal"
                      class="w-3.5 h-3.5 text-primary"
                    />
                    Automated Onboarding Command (CLI)
                  </label>
                  <button
                    type="button"
                    class="text-xs text-primary hover:text-primary-700 flex items-center gap-1 font-medium focus:outline-none"
                    @click="copyText(generatedCliCommand, 'CLI Command')"
                  >
                    <UIcon
                      name="i-lucide-copy"
                      class="w-3.5 h-3.5"
                    />
                    Copy Command
                  </button>
                </div>
                <div class="bg-gray-950 text-gray-200 p-3.5 rounded-xl font-mono text-xs overflow-x-auto shadow-inner border border-gray-800">
                  <code>{{ generatedCliCommand }}</code>
                </div>
              </div>

              <!-- Provisioning Progress Simulator (when executing) -->
              <div
                v-if="isProvisioning"
                class="p-4 rounded-xl bg-gray-900 text-white space-y-3 font-mono text-xs border border-primary/30"
              >
                <div class="flex items-center justify-between text-xs">
                  <span class="text-primary-400 flex items-center gap-2">
                    <UIcon
                      name="i-lucide-loader-2"
                      class="w-4 h-4 animate-spin"
                    />
                    Provisioning {{ targetDomain }}...
                  </span>
                  <span class="text-gray-400">{{ provisioningProgress }}%</span>
                </div>
                <div class="w-full bg-gray-800 h-2 rounded-full overflow-hidden">
                  <div
                    class="bg-primary h-full transition-all duration-300 rounded-full"
                    :style="{ width: `${provisioningProgress}%` }"
                  />
                </div>
                <div class="space-y-1 text-[11px] text-gray-300 max-h-32 overflow-y-auto pt-1">
                  <div
                    v-for="(log, i) in provisioningLogs"
                    :key="i"
                    class="flex items-center gap-2"
                  >
                    <span class="text-emerald-400 font-bold">✓</span>
                    <span>{{ log }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Card Footer Controls -->
          <div class="px-6 sm:px-8 py-4 border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-850 flex items-center justify-between">
            <button
              v-if="currentStep > 1"
              type="button"
              :disabled="isProvisioning"
              class="px-4 py-2 rounded-xl text-xs font-semibold text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              @click="currentStep--"
            >
              Back
            </button>
            <div v-else />

            <div class="flex items-center gap-2">
              <ReusableButton
                v-if="currentStep < steps.length"
                variant="fill"
                color="primary"
                size="sm"
                :disabled="!isCurrentStepValid"
                @click="currentStep++"
              >
                Continue
              </ReusableButton>

              <ReusableButton
                v-else
                variant="fill"
                color="primary"
                size="sm"
                :loading="isProvisioning"
                @click="handleProvisionSite"
              >
                Generate & Deploy Site
              </ReusableButton>
            </div>
          </div>
        </template>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch } from 'vue'
import AppFormField from '~/components/shared/AppFormField.vue'
import ReusableButton from '~/components/shared/ReusableButton.vue'
import ReusableSelectMenu from '~/components/shared/ReusableSelectMenu.vue'
import Logo from '~/components/Logo.vue'
import { useAppToast } from '~/composables/useAppToast'

definePageMeta({
  layout: false
})

const { success } = useAppToast()

interface DeployedSite {
  clientName: string
  slug: string
  domain: string
  apiDomain: string
  industry: string
  brandColor: string
  adminName: string
  adminEmail: string
  frontendPort: number
  kongPort: number
  complianceFramework: string
}

const currentStep = ref(1)
const isProvisioning = ref(false)
const provisioningProgress = ref(0)
const provisioningLogs = ref<string[]>([])
const isDeploymentComplete = ref(false)
const deployedSite = ref<DeployedSite | null>(null)

const steps = [
  { id: 1, title: 'Organization' },
  { id: 2, title: 'Administrator' },
  { id: 3, title: 'Infrastructure' },
  { id: 4, title: 'Review & Deploy' }
]

const industryOptions = [
  'Consulting & Professional Services',
  'Telecommunications',
  'Banking & Financial Services',
  'Energy, Oil & Gas',
  'State-Owned Enterprise (BUMN)',
  'Manufacturing & Supply Chain',
  'Healthcare & Pharmaceuticals',
  'Technology & Software'
]

const frameworkOptions = [
  'ISO 27001 + IIA Global Standards',
  'COSO ERM Framework',
  'Bank Indonesia / OJK Regulatory Standard',
  'UU PDP (Undang-Undang Perlindungan Data Pribadi)',
  'NIST Cybersecurity Framework'
]

const form = reactive({
  clientName: '',
  slug: '',
  industry: 'Consulting & Professional Services',
  brandColor: '#0284c7',
  complianceFramework: 'ISO 27001 + IIA Global Standards',
  adminName: '',
  adminEmail: '',
  adminPhone: '',
  enforceMfa: true,
  autoConfigInfrastructure: true,
  gdriveFolderId: '1bX7yZ9kL0mN8pQ2rS4tU6vW8xYz1234A',
  serverIp: '202.10.34.166',
  frontendPort: 3014,
  kongPort: 8094
})

const isSlugTouched = ref(false)

// Auto-generate slug from Client Name if user hasn't manually edited it
watch(() => form.clientName, (name) => {
  if (!isSlugTouched.value) {
    form.slug = name
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]/g, '')
      .slice(0, 30)
  }
})

const sanitizeSlug = () => {
  isSlugTouched.value = true
  form.slug = form.slug.toLowerCase().replace(/[^a-z0-9-]/g, '')
}

const targetDomain = computed(() => {
  return `${form.slug || 'client'}.auditsphere.id`
})

const targetApiDomain = computed(() => {
  return `api-${form.slug || 'client'}.auditsphere.id`
})

const generatedDatabases = computed(() => {
  const s = form.slug || 'tenant'
  return [
    `rb_audit_auth_${s}`,
    `rb_audit_audit_${s}`,
    `rb_audit_master_${s}`,
    `rb_audit_risk_${s}`,
    `rb_audit_analytics_${s}`
  ]
})

const generatedCliCommand = computed(() => {
  return `./scripts/onboard-tenant.sh ${form.slug || 'slug'} "${form.clientName || 'Client'}" "${form.gdriveFolderId || 'GDRIVE_ID'}" ${form.frontendPort} ${form.kongPort}`
})

const isCurrentStepValid = computed(() => {
  if (currentStep.value === 1) {
    return form.clientName.trim().length >= 2 && form.slug.trim().length >= 2 && !!form.industry
  }
  if (currentStep.value === 2) {
    return form.adminName.trim().length >= 2 && form.adminEmail.includes('@')
  }
  if (currentStep.value === 3) {
    if (form.autoConfigInfrastructure) return true
    return form.frontendPort > 1000 && form.kongPort > 1000 && form.gdriveFolderId.length > 5
  }
  return true
})

const copyText = async (text: string, label: string) => {
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    }
    success('Copied to Clipboard', `${label} copied successfully.`)
  } catch {
    success('Copied', text)
  }
}

const handleProvisionSite = async () => {
  isProvisioning.value = true
  provisioningProgress.value = 10
  provisioningLogs.value = ['Resolving DNS *.auditsphere.id for ' + targetDomain.value]

  await new Promise(r => setTimeout(r, 600))
  provisioningProgress.value = 35
  provisioningLogs.value.push('Created PostgreSQL isolated databases (rb_audit_*_' + form.slug + ')')

  await new Promise(r => setTimeout(r, 700))
  provisioningProgress.value = 65
  provisioningLogs.value.push('Executed schema migrations & seeded default RBAC roles')

  await new Promise(r => setTimeout(r, 600))
  provisioningProgress.value = 85
  provisioningLogs.value.push('Configured Nginx reverse proxy & Kong Gateway port ' + form.kongPort)

  await new Promise(r => setTimeout(r, 500))
  provisioningProgress.value = 100
  provisioningLogs.value.push('SSL Let\'s Encrypt certificate activated for ' + targetDomain.value)

  await new Promise(r => setTimeout(r, 400))
  isProvisioning.value = false

  deployedSite.value = {
    clientName: form.clientName,
    slug: form.slug,
    domain: targetDomain.value,
    apiDomain: targetApiDomain.value,
    industry: form.industry,
    brandColor: form.brandColor,
    adminName: form.adminName,
    adminEmail: form.adminEmail,
    frontendPort: form.frontendPort,
    kongPort: form.kongPort,
    complianceFramework: form.complianceFramework
  }

  isDeploymentComplete.value = true
  success('Client Site Provisioned', `${form.clientName} (${targetDomain.value}) is now live.`)
}

const resetForm = () => {
  form.clientName = ''
  form.slug = ''
  form.industry = 'Consulting & Professional Services'
  form.brandColor = '#0284c7'
  form.complianceFramework = 'ISO 27001 + IIA Global Standards'
  form.adminName = ''
  form.adminEmail = ''
  form.adminPhone = ''
  form.enforceMfa = true
  form.autoConfigInfrastructure = true
  form.frontendPort += 1
  form.kongPort += 1
  isSlugTouched.value = false
  currentStep.value = 1
  isDeploymentComplete.value = false
  deployedSite.value = null
}
</script>
