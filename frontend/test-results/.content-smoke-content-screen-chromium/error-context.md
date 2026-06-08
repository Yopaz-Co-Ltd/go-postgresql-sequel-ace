# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: .content-smoke.spec.cjs >> content screen
- Location: .content-smoke.spec.cjs:2:1

# Error details

```
Error: locator.click: Error: strict mode violation: getByRole('button', { name: 'Connect' }) resolved to 2 elements:
    1) <button type="button" class="primary connect-button">Connect</button> aka getByRole('button', { name: 'Connect', exact: true })
    2) <button type="button">Test connection</button> aka getByRole('button', { name: 'Test connection' })

Call log:
  - waiting for getByRole('button', { name: 'Connect' })

```

# Page snapshot

```yaml
- main [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - strong [ref=e10]: Sequel Ace
      - generic [ref=e11]:
        - text: Choose Database...
        - img [ref=e12]
      - navigation "Database tools" [ref=e14]:
        - button "Structure" [ref=e15] [cursor=pointer]:
          - img [ref=e16]
          - generic [ref=e22]: Structure
        - button "Content" [ref=e23] [cursor=pointer]:
          - img [ref=e24]
          - generic [ref=e25]: Content
        - button "Relations" [ref=e26] [cursor=pointer]:
          - img [ref=e27]
          - generic [ref=e30]: Relations
        - button "Triggers" [ref=e31] [cursor=pointer]:
          - img [ref=e32]
          - generic [ref=e35]: Triggers
        - button "Table Info" [ref=e36] [cursor=pointer]:
          - img [ref=e37]
          - generic [ref=e39]: Table Info
        - button "Query" [ref=e40] [cursor=pointer]:
          - img [ref=e41]
          - generic [ref=e44]: Query
        - button "Table History" [ref=e45] [cursor=pointer]:
          - img [ref=e46]
          - generic [ref=e49]: Table History
        - button "Users" [ref=e50] [cursor=pointer]:
          - img [ref=e51]
          - generic [ref=e56]: Users
        - button "Console" [ref=e57] [cursor=pointer]:
          - img [ref=e58]
          - generic [ref=e60]: Console
    - complementary [ref=e61]:
      - generic [ref=e62]:
        - img [ref=e63]
        - text: QUICK CONNECT
      - heading "FAVORITES" [level=2] [ref=e65]
      - generic [ref=e66]:
        - button "127.0.0.1" [active] [ref=e67] [cursor=pointer]:
          - img [ref=e68]
          - generic [ref=e72]: 127.0.0.1
        - button "mory-dev" [ref=e73] [cursor=pointer]:
          - img [ref=e74]
          - generic [ref=e78]: mory-dev
        - button "real.reddy.vn" [ref=e79] [cursor=pointer]:
          - img [ref=e80]
          - generic [ref=e84]: real.reddy.vn
        - button "real.chipmunk.vn" [ref=e85] [cursor=pointer]:
          - img [ref=e86]
          - generic [ref=e90]: real.chipmunk.vn
        - button "itech_iot" [ref=e91] [cursor=pointer]:
          - img [ref=e92]
          - generic [ref=e96]: itech_iot
        - button "New Favorite" [ref=e97] [cursor=pointer]:
          - img [ref=e98]
          - generic [ref=e102]: New Favorite
        - button "spispi" [ref=e103] [cursor=pointer]:
          - img [ref=e104]
          - generic [ref=e108]: spispi
        - button "php-dev" [ref=e109] [cursor=pointer]:
          - img [ref=e110]
          - generic [ref=e114]: php-dev
        - button "chipmunk" [ref=e115] [cursor=pointer]:
          - img [ref=e116]
          - generic [ref=e120]: chipmunk
        - button "shopbay.cloud" [ref=e121] [cursor=pointer]:
          - img [ref=e122]
          - generic [ref=e126]: shopbay.cloud
        - button "yopaz.dev" [ref=e127] [cursor=pointer]:
          - img [ref=e128]
          - generic [ref=e132]: yopaz.dev
        - button "127.0.0.1 docker" [ref=e133] [cursor=pointer]:
          - img [ref=e134]
          - generic [ref=e138]: 127.0.0.1 docker
        - button "quanlv" [ref=e139] [cursor=pointer]:
          - img [ref=e140]
          - generic [ref=e144]: quanlv
        - button "goldwin-hub" [ref=e145] [cursor=pointer]:
          - img [ref=e146]
          - generic [ref=e150]: goldwin-hub
        - button "golang-ci" [ref=e151] [cursor=pointer]:
          - img [ref=e152]
          - generic [ref=e156]: golang-ci
        - button "web3" [ref=e157] [cursor=pointer]:
          - img [ref=e158]
          - generic [ref=e162]: web3
        - button "itech" [ref=e163] [cursor=pointer]:
          - img [ref=e164]
          - generic [ref=e168]: itech
        - button "goldwin" [ref=e169] [cursor=pointer]:
          - img [ref=e170]
          - generic [ref=e174]: goldwin
        - button "lumine-v2" [ref=e175] [cursor=pointer]:
          - img [ref=e176]
          - generic [ref=e180]: lumine-v2
        - button "quan.pro.vn" [ref=e181] [cursor=pointer]:
          - img [ref=e182]
          - generic [ref=e186]: quan.pro.vn
        - button "3307" [ref=e187] [cursor=pointer]:
          - img [ref=e188]
          - generic [ref=e192]: "3307"
      - generic [ref=e193]:
        - button "Options" [ref=e194] [cursor=pointer]:
          - img [ref=e195]
        - button "Add folder" [ref=e197] [cursor=pointer]:
          - img [ref=e198]
        - button "Add favorite" [ref=e200] [cursor=pointer]:
          - img [ref=e201]
        - button "Sidebar" [ref=e202] [cursor=pointer]:
          - img [ref=e203]
    - generic [ref=e206]:
      - paragraph [ref=e207]: Enter connection details below, or choose a favorite
      - generic [ref=e208]:
        - generic [ref=e209]:
          - tablist [ref=e210]:
            - button "TCP/IP" [ref=e211] [cursor=pointer]
            - button "Socket" [ref=e212] [cursor=pointer]
            - button "SSH" [ref=e213] [cursor=pointer]
            - button "AWS IAM" [ref=e214] [cursor=pointer]
          - generic [ref=e215]:
            - generic [ref=e216]:
              - generic [ref=e217]: "Name:"
              - textbox "Name:" [ref=e218]: 127.0.0.1
            - generic "Favorite color" [ref=e219]:
              - img [ref=e220]
              - button [ref=e223] [cursor=pointer]
              - button [ref=e224] [cursor=pointer]
              - button [ref=e225] [cursor=pointer]
              - button [ref=e226] [cursor=pointer]
              - button [ref=e227] [cursor=pointer]
              - button [ref=e228] [cursor=pointer]
              - button [ref=e229] [cursor=pointer]
            - generic [ref=e230]:
              - generic [ref=e231]: "Host:"
              - textbox "Host:" [ref=e232]: 127.0.0.1
            - generic [ref=e233]:
              - generic [ref=e234]: "Username:"
              - textbox "Username:" [ref=e235]: postgres
            - generic [ref=e236]:
              - generic [ref=e237]: "Password:"
              - textbox "Password:" [ref=e238]: postgres
            - generic [ref=e239]:
              - generic [ref=e240]: "Database:"
              - textbox "Database:" [ref=e241]:
                - /placeholder: optional
                - text: sample_store
            - generic [ref=e242]:
              - generic [ref=e243]: "Port:"
              - spinbutton "Port:" [ref=e244]: "5432"
            - generic [ref=e245]:
              - generic [ref=e246]: "Time Zone:"
              - combobox "Time Zone:" [ref=e247]:
                - option "Use Server Time Zone" [selected]
                - option "UTC"
                - option "Asia/Ho_Chi_Minh"
            - generic [ref=e248]:
              - checkbox "Allow LOCAL_DATA_INFILE (insecure)" [ref=e249]
              - generic [ref=e250]: Allow LOCAL_DATA_INFILE (insecure)
            - generic [ref=e251]:
              - checkbox "Enable Cleartext plugin (insecure)" [ref=e252]
              - generic [ref=e253]: Enable Cleartext plugin (insecure)
            - generic [ref=e254]:
              - checkbox "Require SSL" [ref=e255]
              - generic [ref=e256]: Require SSL
        - generic [ref=e257]:
          - button "Help" [ref=e258] [cursor=pointer]:
            - img [ref=e259]
          - button "Connect" [ref=e262] [cursor=pointer]
          - button "Add to Favorites" [ref=e263] [cursor=pointer]
          - button "Save changes" [ref=e264] [cursor=pointer]
          - button "Test connection" [ref=e265] [cursor=pointer]
```

# Test source

```ts
  1 | const { test } = require('@playwright/test');
  2 | test('content screen', async ({ page }) => {
  3 |   await page.goto('http://localhost:5173/');
  4 |   await page.getByRole('button', { name: '127.0.0.1', exact: true }).click();
> 5 |   await page.getByRole('button', { name: 'Connect' }).click();
    |                                                       ^ Error: locator.click: Error: strict mode violation: getByRole('button', { name: 'Connect' }) resolved to 2 elements:
  6 |   await page.waitForSelector('.content-grid table', { timeout: 10000 });
  7 |   await page.screenshot({ path: '/tmp/sequel-content.png', fullPage: true });
  8 | });
  9 | 
```