
pipeline {
    agent any

    environment {
        GOPATH = "${WORKSPACE}"
        GO111MODULE = "on"
        PROJ_DIR = "${WORKSPACE}/src/github.com/33cn/plugin"
    }

    options {
        timeout(time: 2,unit: 'HOURS')
        retry(1)
        timestamps()
        gitLabConnection('gitlab33')
        gitlabBuilds(builds: ['check'])
        checkoutToSubdirectory "src/github.com/33cn/plugin"
    }
    tools {go 'go'}
    stages {
        stage('deploy') {
            steps {
                dir("${PROJ_DIR}"){
                    gitlabCommitStatus(name: 'deploy'){
                        sh '''
                            set -e
                            go version
                            make build_ci
                            cd build
                            rm -rf "${BUILD_NUMBER}"
                            mkdir -p "${BUILD_NUMBER}"
                            cp -r ci/* "${BUILD_NUMBER}/"
                            ./docker-compose-pre.sh modify
                            cp chain33* Dockerfile* docker-compose* *.sh "${BUILD_NUMBER}/"
                            cd "${BUILD_NUMBER}"
                            rm -rf cross2eth mix rgbx rollup|| true
                            ./docker-compose-pre.sh run "${BUILD_NUMBER}" all
                        '''
                    }
                }
            }

            post {
                always {
                    dir("${PROJ_DIR}"){
                        sh '''
                            set +e
                            cd build
                            if [ -d "${BUILD_NUMBER}" ]; then
                                cd "${BUILD_NUMBER}"
                                ./docker-compose-pre.sh down "${BUILD_NUMBER}" all || true
                                cd ..
                                rm -rf "${BUILD_NUMBER}"
                            fi
                            cd ..
                            make clean || true
                        '''
                    }
                }
            }
        }
    }

    post {
        always {
            echo 'One way or another, I have finished'
            // clean up our workspace
            deleteDir()
        }

        success {
            echo 'I succeeeded!'
            echo "email user: ${ghprbActualCommitAuthorEmail}"

            script{
                try {
                    mail to: "${ghprbActualCommitAuthorEmail}",
                         subject: "Successed Pipeline: ${currentBuild.fullDisplayName}",
                         body: "this is success with ${env.BUILD_URL}"
                }
                catch (err){
                    echo 'email  err'
                }
                currentBuild.result = 'SUCCESS'
            }
            echo 'SUCCESS'

        }

        failure {
            echo 'I failed '
            echo "email user: ${ghprbActualCommitAuthorEmail}"
            script{
                try {
                    mail to: "${ghprbActualCommitAuthorEmail}",
                         subject: "Failed Pipeline: ${currentBuild.fullDisplayName}",
                         body: "Something is wrong with ${env.BUILD_URL}"
                }catch (err){
                    echo 'email err'
                }
            }

            echo currentBuild.result
        }
    }
}
