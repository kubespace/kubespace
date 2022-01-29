<template>
    <div>
        <div :id=id style="width: 100%; height: 300px;"></div>
    </div>
</template>

<script>
    import Highcharts from 'highcharts';
    import HighchartsMore from 'highcharts/highcharts-more';
    import SolidGauge from 'highcharts/modules/solid-gauge.js'
    // import resize from '@/components/charts/mixins/resize'

    HighchartsMore(Highcharts)
    SolidGauge(Highcharts);

    if (!Highcharts.theme) {
        Highcharts.setOptions({
            chart: {
                backgroundColor: 'white'
            },
            colors: ['#F62366', '#9DFF02', '#0CCDD6'],
            title: {
                style: {
                    color: 'silver'
                }
            },
            tooltip: {
                style: {
                    color: 'silver'
                }
            }
        });
    }

    export default {
        // mixins: [resize],
        mounted() {
            this.init();

        },
        data() {
            return {

            }
        },
        props: {
            id: {
                type: String,
                default: "charts"
            },
            data: {
                type: Object,
                default: function () {
                    return {
                        backgroundNum: 0,
                        series: []
                    }
                }
            }
        }, 
        methods: {
            setBackground() {
                let arr = []

                let obj = { // Track for Move
                    outerRadius: '',
                    innerRadius: '',
                    backgroundColor: '',
                    borderWidth: 0
                }
                for (let i = 0; i < this.data.backgroundNum; i++) {
                    obj = JSON.parse(JSON.stringify(obj))
                    obj.outerRadius = `${112 - 25 * i}%`
                    obj.innerRadius = `${88 - 25 * i}%`
                    obj.backgroundColor = Highcharts.Color(Highcharts.getOptions().colors[i]).setOpacity(0.3).get()
                    arr.push(obj)
                }
                return arr
            },
            setSeries() {
                let arr = []
                let obj = {
                    name: '',
                    // borderColor: Highcharts.getOptions().colors[1],
                    borderColor: '',
                    data: [{
                        color: Highcharts.getOptions().colors[0],
                        // color:'yellow',
                        radius: '100%',
                        innerRadius: '100%',
                        y: ""
                    }]
                }
                this.data.series.forEach((item, i) => {
                    obj = JSON.parse(JSON.stringify(obj))
                    obj.name = item.name
                    obj.borderColor = Highcharts.getOptions().colors[i]
                    obj.data[0].color = Highcharts.getOptions().colors[i]
                    obj.data[0].y = item.value
                    obj.data[0].radius = `${100 - 25 * i}%`
                    obj.data[0].innerRadius = `${100 - 25 * i}%`
                    arr.push(obj)
                })
                return arr
            },
            init() {
                console.log(12312)
                this.draw();
            },
            draw() {
                new Highcharts.chart(this.id, {
                    chart: {
                        type: 'solidgauge',
                        marginTop: 50
                    },
                    credits: { enabled: false },
                    title: {
                        text: this.data.title,
                        style: {
                            fontSize: '18px'
                        }
                    },
                    tooltip: {
                        borderWidth: 0,
                        backgroundColor: 'none',
                        shadow: false,
                        style: {
                            fontSize: '10px'
                        },
                        pointFormat: '{series.name}<br><span style="font-size:1em; color: {point.color}; font-weight: bold">{point.y}%</span>',
                        positioner: function (labelWidth) {
                            return {
                                x: 94 - labelWidth / 2,
                                y: 140
                            };
                        }
                    },
                    pane: {
                        startAngle: 0,
                        endAngle: 360,
                        background: this.setBackground(),
                    },
                    yAxis: {
                        min: 0,
                        max: 100,
                        lineWidth: 0,
                        tickPositions: []
                    },
                    plotOptions: {
                        solidgauge: {
                            borderWidth: '17px', //覆盖层 宽度
                            dataLabels: {
                                enabled: false
                            },
                            linecap: 'round',
                            stickyTracking: false
                        }
                    },
                    series: this.setSeries(),
                });
            }
        }
    }
</script>

<style lang="stylus">

</style>