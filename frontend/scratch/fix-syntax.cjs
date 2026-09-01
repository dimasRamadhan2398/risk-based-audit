const fs = require('fs');
const path = require('path');

const directory = 'c:/Users/Dimas/risk-based-audit/frontend';

function walk(dir) {
    let results = [];
    const list = fs.readdirSync(dir);
    list.forEach(file => {
        file = path.join(dir, file);
        const stat = fs.statSync(file);
        if (stat && stat.isDirectory()) {
            if (!file.includes('node_modules') && !file.includes('.nuxt') && !file.includes('.output')) {
                results = results.concat(walk(file));
            }
        } else {
            if (file.endsWith('.ts') || file.endsWith('.vue')) {
                results.push(file);
            }
        }
    });
    return results;
}

const files = walk(directory);

files.forEach(file => {
    let content = fs.readFileSync(file, 'utf8');
    let changed = false;
    
    // Fix { description: t('...' }))) -> { description: t('...') }))
    if (content.includes("useGlobalModalStore().confirmDelete({ description: t")) {
        let newContent = content.replace(/\{ description:\s*t\((['"][^'"]+['"](?:,\s*\{[^}]*\})?)\s*\}\)\)\)/g, "{ description: t($1) }))");
        if (newContent !== content) {
            content = newContent;
            changed = true;
        }
    }
    
    // Also there are some confirm(`...`)?
    // Let's check for template literals missing a closing backtick or paren
    // If original was confirm(`Are you sure you want to delete employee "${employee.full_name}"?`)
    // The regex caught `Are you sure you want to delete employee "${employee.full_name}"?`
    // And it became { description: `...` } which is fine.
    
    // There are some errors in stores/activity-plan.ts at line 307 - maybe extra `};`?
    
    if (changed) {
        fs.writeFileSync(file, content, 'utf8');
        console.log(`Fixed t() parenthesis in ${file}`);
    }
});
