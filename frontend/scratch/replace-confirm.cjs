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

let replaceCount = 0;

files.forEach(file => {
    let content = fs.readFileSync(file, 'utf8');
    
    // We want to replace `confirm(...)` but NOT `window.confirm(...)` if it exists.
    // Let's just match `confirm(...)` where it's a standalone function call.
    if (/\bconfirm\s*\(/.test(content)) {
        let changed = false;
        
        let lines = content.split('\n');
        for (let i = 0; i < lines.length; i++) {
            let line = lines[i];
            if (/\bconfirm\s*\(/.test(line)) {
                // Find the nearest enclosing function going upwards and make it async if it's not already
                let j = i;
                let foundFunc = false;
                while (j >= 0 && j >= i - 50) {
                    let lookLine = lines[j];
                    // Match `function foo(` or `const foo = (` or `foo() {`
                    if (/(?:function\s+\w+\s*\(|\w+\s*:\s*(?:function)?\s*\(|\w+\s*=\s*(?:async\s*)?(?:\([^)]*\)|\w+)\s*=>|\b\w+\s*\([^)]*\)\s*\{)/.test(lookLine)) {
                        if (!/\basync\b/.test(lookLine) && !lookLine.includes('defineStore') && !lookLine.includes('defineProps') && !lookLine.includes('defineEmits')) {
                            // Let's inject async
                            if (/\bfunction\b/.test(lookLine)) {
                                lines[j] = lookLine.replace(/\bfunction\b/, 'async function');
                            } else if (/=>/.test(lookLine)) {
                                lines[j] = lookLine.replace(/(\([^)]*\)|\w+)\s*=>/, 'async $1 =>');
                            } else if (/\b\w+\s*\([^)]*\)\s*\{/.test(lookLine)) {
                                lines[j] = lookLine.replace(/(\b\w+\s*\([^)]*\)\s*\{)/, 'async $1');
                            }
                        }
                        foundFunc = true;
                        break;
                    }
                    j--;
                }
                
                // Replace confirm(MSG) with await useGlobalModalStore().confirmDelete({ description: MSG })
                lines[i] = line.replace(/\bconfirm\s*\((.*?)\)/, (match, p1) => {
                    return `await useGlobalModalStore().confirmDelete({ description: ${p1} })`;
                });
                changed = true;
                replaceCount++;
            }
        }
        
        if (changed) {
            fs.writeFileSync(file, lines.join('\n'), 'utf8');
            console.log(`Updated ${file}`);
        }
    }
});

console.log(`Replaced ${replaceCount} occurrences.`);
