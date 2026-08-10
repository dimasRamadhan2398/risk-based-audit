<template>
    <UCard :ui="{ header: 'flex flex-col gap-4', body: 'p-0' }">
      <template #header>
        <div class="justify-start text-left">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('settings.faq.title') }}</h3>
            <h6 class="text-sm tracking-wider text-gray-500 dark:text-gray-400">{{ t('settings.faq.subtitle') }}</h6>
        </div>
        <div class="justify-start text-left">
            <UInput v-model="searchQuery" icon="i-lucide-search" size="xl" variant="outline" :placeholder="t('settings.faq.searchPlaceholder')" :ui="{
                root: ''
            }"></UInput>

            <div class="flex flex-wrap gap-2 mt-3">
                <p class="text-sm font-medium text-neutral-700 dark:text-neutral-300 self-center">{{ t('settings.faq.categoryLabel') }}</p>
                <UBadge
                    v-for="category in rawCategories"
                    :key="category"
                    :color="selectedCategory === category ? 'primary' : 'neutral'"
                    :variant="selectedCategory === category ? 'solid' : 'subtle'"
                    class="cursor-pointer hover:opacity-80 transition-opacity"
                    @click="selectedCategory = category"
                >
                    {{ t(`settings.faq.categories.${category}`) }}
                </UBadge>
            </div>
        </div>
      </template>

      <div class="p-4">
        <UAccordion :items="filteredFaqs" multiple variant="subtle" color="primary">
          <template #default="{ item, open }">
            <UButton color="neutral" variant="ghost" class="w-full justify-between">                                                                                                                      
              <span class="font-medium ">{{ item.label }}</span>                                                                                                                      
              <UIcon :name="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-5" />
            </UButton>
          </template>
          <template #content="{ item }">
            <p class="text-sm text-gray-600 dark:text-gray-300 mt-2">{{ item.content }}</p>
          </template>
        </UAccordion>
      </div>
    </UCard>
</template>
<script setup>
const { t } = useI18n()

const searchQuery = ref('')
const selectedCategory = ref('All')

const rawCategories = [
    'All',
    'General',
    'Account',
    'Security',
    'Audit',
    'Risk'
]

const faqs = computed(() => [
  {
    label: t('settings.faq.q1.question'),
    content: t('settings.faq.q1.answer'),
    category: 'Security'
  },
  {
    label: t('settings.faq.q2.question'),
    content: t('settings.faq.q2.answer'),
    category: 'Risk'
  },
  {
    label: t('settings.faq.q3.question'),
    content: t('settings.faq.q3.answer'),
    category: 'Audit'
  },
  {
    label: t('settings.faq.q4.question'),
    content: t('settings.faq.q4.answer'),
    category: 'General'
  },
  {
    label: t('settings.faq.q5.question'),
    content: t('settings.faq.q5.answer'),
    category: 'Risk'
  },
])

const filteredFaqs = computed(() => {
    let result = faqs.value

    // Filter by search query (label)
    if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase()
        result = result.filter(faq =>
            faq.label.toLowerCase().includes(query)
        )
    }

    // Filter by category
    if (selectedCategory.value !== 'All') {
        result = result.filter(faq => {
            return faq.category === selectedCategory.value
        })
    }

    return result
})
</script>