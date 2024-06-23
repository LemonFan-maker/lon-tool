$url="https://git.timoxa0.su/timoxa0/lon-tool/releases/download/latest/lon-tool_win_amd64.exe"
$bin_dir = Join-Path $env:USERPROFILE ".bin"
$platform_tools_url = "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
$platform_tools_dir = Join-Path $bin_dir "platform_tools"

function Get-FileFromWeb {
    param (
        # Parameter help description
        [Parameter(Mandatory)]
        [string]$URL,
  
        # Parameter help description
        [Parameter(Mandatory)]
        [string]$File 
    )
    Begin {
        function Show-Progress {
            param (
                # Enter total value
                [Parameter(Mandatory)]
                [Single]$TotalValue,
        
                # Enter current value
                [Parameter(Mandatory)]
                [Single]$CurrentValue,
        
                # Enter custom progresstext
                [Parameter(Mandatory)]
                [string]$ProgressText,
        
                # Enter value suffix
                [Parameter()]
                [string]$ValueSuffix,
        
                # Enter bar lengh suffix
                [Parameter()]
                [int]$BarSize = 40,

                # show complete bar
                [Parameter()]
                [switch]$Complete
            )
            
            # calc %
            $percent = $CurrentValue / $TotalValue
            $percentComplete = $percent * 100
            if ($ValueSuffix) {
                $ValueSuffix = " $ValueSuffix" # add space in front
            }
            if ($psISE) {
                Write-Progress "$ProgressText $CurrentValue$ValueSuffix of $TotalValue$ValueSuffix" -id 0 -percentComplete $percentComplete            
            }
            else {
                # build progressbar with string function
                $curBarSize = $BarSize * $percent
                $progbar = ""
                $progbar = $progbar.PadRight($curBarSize,[char]9608)
                $progbar = $progbar.PadRight($BarSize,[char]9617)
        
                if (!$Complete.IsPresent) {
                    Write-Host -NoNewLine "`r$ProgressText $progbar [ $($CurrentValue.ToString("#.###").PadLeft($TotalValue.ToString("#.###").Length))$ValueSuffix / $($TotalValue.ToString("#.###"))$ValueSuffix ] $($percentComplete.ToString("##0.00").PadLeft(6)) % complete"
                }
                else {
                    Write-Host -NoNewLine "`r$ProgressText $progbar [ $($TotalValue.ToString("#.###").PadLeft($TotalValue.ToString("#.###").Length))$ValueSuffix / $($TotalValue.ToString("#.###"))$ValueSuffix ] $($percentComplete.ToString("##0.00").PadLeft(6)) % complete"                    
                }                
            }   
        }
    }
    Process {
        try {
            $storeEAP = $ErrorActionPreference
            $ErrorActionPreference = 'Stop'
        
            # invoke request
            $request = [System.Net.HttpWebRequest]::Create($URL)
            $response = $request.GetResponse()
  
            if ($response.StatusCode -eq 401 -or $response.StatusCode -eq 403 -or $response.StatusCode -eq 404) {
                throw "Remote file either doesn't exist, is unauthorized, or is forbidden for '$URL'."
            }
  
            if($File -match '^\.\\') {
                $File = Join-Path (Get-Location -PSProvider "FileSystem") ($File -Split '^\.')[1]
            }
            
            if($File -and !(Split-Path $File)) {
                $File = Join-Path (Get-Location -PSProvider "FileSystem") $File
            }

            if ($File) {
                $fileDirectory = $([System.IO.Path]::GetDirectoryName($File))
                if (!(Test-Path($fileDirectory))) {
                    [System.IO.Directory]::CreateDirectory($fileDirectory) | Out-Null
                }
            }

            [long]$fullSize = $response.ContentLength
            $fullSizeMB = $fullSize / 1024 / 1024
  
            # define buffer
            [byte[]]$buffer = new-object byte[] 1048576
            [long]$total = [long]$count = 0
  
            # create reader / writer
            $reader = $response.GetResponseStream()
            $writer = new-object System.IO.FileStream $File, "Create"
  
            # start download
            $finalBarCount = 0 #show final bar only one time
            do {
          
                $count = $reader.Read($buffer, 0, $buffer.Length)
          
                $writer.Write($buffer, 0, $count)
              
                $total += $count
                $totalMB = $total / 1024 / 1024
          
                if ($fullSize -gt 0) {
                    Show-Progress -TotalValue $fullSizeMB -CurrentValue $totalMB -ProgressText "Downloading $($File.Name)" -ValueSuffix "MB"
                }

                if ($total -eq $fullSize -and $count -eq 0 -and $finalBarCount -eq 0) {
                    Show-Progress -TotalValue $fullSizeMB -CurrentValue $totalMB -ProgressText "Downloading $($File.Name)" -ValueSuffix "MB" -Complete
                    $finalBarCount++
                    #Write-Host "$finalBarCount"
                }

            } while ($count -gt 0)
        }
  
        catch {
        
            $ExeptionMsg = $_.Exception.Message
            Write-Host "Download breaks with error : $ExeptionMsg"
        }
  
        finally {
            # cleanup
            if ($reader) { $reader.Close() }
            if ($writer) { $writer.Flush(); $writer.Close() }
        
            $ErrorActionPreference = $storeEAP
            [GC]::Collect()
        }    
    }
}

function Install-Tool {
    if (-not (Test-Path $bin_dir -PathType Container)) {
        New-Item -Path $bin_dir -ItemType Directory | Out-Null
    }
    
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User") -split ";"
    if ($currentPath -notcontains $bin_dir) {
        [Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$bin_dir", "User")
        $env:PATH="$env:PATH;$bin_dir"
        Write-Host "$bin_dir added to PATH."
    }
    
    Get-FileFromWeb "$url" (Join-Path $bin_dir "lon-tool.exe")
    Write-Host
}

function Install-Platoform_tools {
    
    if (-not (Test-Path $platform_tools_dir -PathType Container)) {
        New-Item -Path $platform_tools_dir -ItemType Directory | Out-Null
    }
    
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User") -split ";"
    if ($currentPath -notcontains $platform_tools_dir) {
        [Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$platform_tools_dir", "User")
        $env:PATH="$env:PATH;$platform_tools_dir"
        Write-Host "$platform_tools_dir added to PATH."
    }
    
    Get-FileFromWeb "$platform_tools_url" (Join-Path $platform_tools_dir "tools.zip")
    Write-Host
    Expand-Archive -Path (Join-Path $platform_tools_dir "tools.zip") -DestinationPath $platform_tools_dir
    Move-Item (Join-Path $platform_tools_dir "platform-tools\*") $platform_tools_dir
    Remove-Item (Join-Path $platform_tools_dir "platform-tools\*")
}

Install-Tool

if (-not (Get-Command "adb.exe" -ErrorAction SilentlyContinue)) {
    $decision = $Host.UI.PromptForChoice("Adb executable not found in PATH", "Do you wand to install android platform tools?", ("&Yes", "&No"), 1)
    if ($decision -eq 0) {
        Install-Platoform_tools
    }
}
Write-Host "Done!" -ForegroundColor Green
