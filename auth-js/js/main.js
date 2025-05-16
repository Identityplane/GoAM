import { addAndDouble } from './lib/math.js'
import { initEmailOTP } from './lib/nodes/emailOTP.js'
import { initPasswordOrSocialLogin } from './lib/nodes/passwordOrSocialLogin.js'
import { initHcaptcha } from './lib/nodes/hcaptcha.js'
import { initVerifyPasskey } from './lib/nodes/verifyPasskey.js'
import { initRegisterPasskey } from './lib/nodes/registerPasskey.js'
// Node-specific initialization

const nodeHandlers = {
    'emailOTP': initEmailOTP,
    'passwordOrSocialLogin': initPasswordOrSocialLogin,
    'hcaptcha': initHcaptcha,

    'verifyPasskey': initVerifyPasskey,
    'registerPasskey': initRegisterPasskey
    // Add more node handlers here as needed
}

// Initialize based on the current node
document.addEventListener('DOMContentLoaded', function () {

    const mainContent = document.querySelector('.main-content')
    if (mainContent) {
        const nodeName = mainContent.dataset.node
        const handler = nodeHandlers[nodeName]
        if (handler) {
            handler()
            console.log('Initialized node:', nodeName)
        }
    }
})