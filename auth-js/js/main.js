import { addAndDouble } from './lib/math.js'
import { initOTP } from './lib/nodes/otp.js'
import { initPasswordOrSocialLogin } from './lib/nodes/passwordOrSocialLogin.js'
import { initHcaptcha } from './lib/nodes/hcaptcha.js'
// Node-specific initialization

const nodeHandlers = {
    'emailOTP': initOTP,
    'passwordOrSocialLogin': initPasswordOrSocialLogin,
    'hcaptcha': initHcaptcha
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